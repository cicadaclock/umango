//! C interface of the race engine, which the Umango application loads at run time.

use std::any::Any;
use std::num::NonZero;
use std::panic::{AssertUnwindSafe, catch_unwind};
use std::path::Path;
use std::sync::Mutex;
use std::thread;

use uma_race_sim::api::{Frame, HorseInput, RaceInput, SkillInput, event_kind, span_kind};
use uma_race_sim::data::{DataEra, GameData, corpus};
use uma_race_sim::sim::defs::Phase;
use uma_race_sim::sim::physics::{hp, stats};
use uma_race_sim::sim::race::{self, DT, Race};
use uma_race_sim::sim::skill::{self, SkillManager};
use uma_race_sim::sim::tick;

/// Version of this interface, checked by the application when it loads the library.
const INTERFACE_VERSION: u32 = 6;

/// HP that runs out within this many meters of the goal does not count as running out short of the goal.
const HP_OUT_GOAL_MARGIN: f32 = 0.1;

/// Most runners in one race.
const MAX_RUNNERS: usize = 18;

/// Lengths per second of the time between two finishers.
const LENGTHS_PER_SECOND: f32 = 6.0;

/// Upper limit in lengths of each winning margin that scores, from Nose upward. A wider margin is a Distance.
const MARGIN_LENGTH_LIMITS: [f32; 20] =
    [0.08, 0.16, 0.32, 0.5, 0.75, 1.0, 1.25, 1.5, 1.75, 2.0, 2.5, 3.0, 3.5, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0];

/// Start delay in seconds below which a start counts as a strong start.
const STRONG_START_DELAY: f32 = 0.02;

/// Target time brackets, each 0.1 second wide, of which the last holds every faster time.
const LAST_TARGET_TIME_BRACKET: i32 = 20;

/// Rarity of a normal skill and of a rare skill. Higher rarities are unique skills.
const RARITY_NORMAL: i32 = 1;
const RARITY_RARE: i32 = 2;
const RARITY_FIRST_UNIQUE: i32 = 3;

/// Steps between two stops of a shared race. Closer stops skip more of each skill race and cost more to keep.
const STOP_INTERVAL_STEPS: u32 = 64;

/// Set in an activation flag when the skill activated.
const FIRED_SKILL: u8 = 1;

/// Set in an activation flag when the lower tier that the skill would replace activated.
const FIRED_REPLACED_TIER: u8 = 2;

/// Text of the last failure, read through `racesim_last_error`.
static LAST_ERROR: Mutex<String> = Mutex::new(String::new());

/// What one runner did in one simulated race. The application holds the same layout.
#[repr(C)]
#[derive(Clone, Copy, Default)]
pub struct RunnerOutcome {
    /// Finish order, counted from 0.
    finish_order: i32,
    /// Finish time in seconds of race simulation time, not the time that the game displays.
    finish_time: f32,
    /// Meters short of the goal when the previous finisher finished, 0 for the winner.
    finish_gap: f32,
    /// Start delay in seconds.
    start_delay: f32,
    /// 1 when the start delay cost the first frame of acceleration, otherwise 0.
    late_start: i32,
    /// Meters from the start of the late race to the start of the last spurt, negative without a last spurt.
    spurt_delay: f32,
    /// Target speed of the last spurt in meters per second.
    spurt_speed: f32,
    /// HP at the start of the race.
    start_hp: f32,
    /// HP at the finish.
    final_hp: f32,
    /// Meters short of the goal when HP first ran out, 0 when HP lasted to the goal.
    hp_out_distance: f32,
    /// HP that the runner needed after HP ran out.
    hp_shortfall: f32,
    /// Seconds spent dueling.
    duel_time: f32,
    /// Seconds in downhill mode before the late race.
    downhill_time_early: f32,
    /// Seconds in downhill mode in the late race.
    downhill_time_late: f32,
    /// Seconds in the Pace Up, Pace Up EX, Speed Up and Overtake modes.
    pace_up_time: f32,
    /// Seconds in Pace Down mode.
    pace_down_time: f32,
    /// Meters lost to lane changes and outside lanes on corners.
    world_transform_loss: f32,
}

/// A stretch of one race in which one runner had one skill active. The application holds the same layout.
#[repr(C)]
#[derive(Clone, Copy)]
pub struct SkillSpan {
    /// Index of the race in the seed list.
    race: i32,
    /// Gate index of the runner.
    runner: i32,
    skill_id: i32,
    /// 1 when the skill has no time limit and lasts the whole race, otherwise 0.
    passive: i32,
    /// Meters from the start at which the skill activated.
    start_distance: f32,
    /// Meters from the start at which the effect ended, at most the course distance.
    end_distance: f32,
}

/// What `racesim_run` stores. The application supplies `outcomes`. The library allocates `skill_spans`, which
/// `racesim_release_skill_spans` frees.
#[repr(C)]
pub struct TrialOutput {
    outcomes: *mut RunnerOutcome,
    outcome_count: usize,
    skill_spans: *mut SkillSpan,
    skill_span_count: usize,
}

/// The seed, conditions and gates of one race of a skill trial. The application holds the same layout.
#[repr(C)]
#[derive(Clone, Copy)]
pub struct RaceVariation {
    seed: i32,
    /// 1 Spring, 2 Summer, 3 Fall, 4 Winter, 5 Cherry Blossom.
    season: i32,
    /// 1 Sunny, 2 Cloudy, 3 Rainy, 4 Snow.
    weather: i32,
    /// 1 Good, 2 Soft, 3 Hard, 4 Bad.
    ground_condition: i32,
    /// Index in the saved race of the runner at each gate.
    gate_order: [u8; MAX_RUNNERS],
}

/// The base scores of a team trial race, from team_stadium_raw_score. The application holds the same layout.
#[repr(C)]
pub struct ScoreRules {
    /// By finish order, counted from 0.
    place: [i32; 12],
    /// By the index in `MARGIN_LENGTH_LIMITS` of the winning margin, then Distance.
    margin: [i32; 21],
    /// By the number of tenths of a second under the standard course time.
    target_time: [i32; 21],
    /// By rarity from the first unique rarity, then by skill level from 1.
    unique_skill: [[i32; 10]; 3],
    rare_skill: i32,
    normal_skill: i32,
    long_shot: i32,
    /// A long shot finishes within this place, counted from 1.
    long_shot_place: i32,
    /// A long shot is this favorite or lower.
    long_shot_popularity: i32,
    rushed: i32,
    rushed_per_second: i32,
    strong_start: i32,
    /// Ahead of all runners of the same running style going into the mid-race and going into the late race.
    positioning: [i32; 2],
}

/// The runner, course, skills and races of a skill trial. The application holds the same layout.
#[repr(C)]
pub struct SkillTrial {
    /// Index in the saved race of the runner under test.
    runner: usize,
    course_set_id: i32,
    race_instance_id: i32,
    /// In ascending order.
    skill_ids: *const i32,
    /// For each skill, the lower tier that the runner owns and the skill would replace, otherwise 0.
    replaced_skill_ids: *const i32,
    skill_count: usize,
    variations: *const RaceVariation,
    variation_count: usize,
    score_rules: *const ScoreRules,
    /// One per variation, written by `racesim_watch_skills` and read by `racesim_run_skill_trial`.
    baselines: *mut BaselineOutcome,
    /// Activation flags, `skill_count` per variation, written by `racesim_watch_skills` and read by
    /// `racesim_run_skill_trial`.
    activations: *mut u8,
}

/// What the runner under test did in one race with no added skill. The application holds the same layout.
#[repr(C)]
#[derive(Clone, Copy, Default)]
pub struct BaselineOutcome {
    /// Team trial score before bonuses.
    score: i32,
    /// Finish order, counted from 0.
    finish_order: i32,
}

/// What one added skill changed in one race, against the baseline of that race. The application holds the same layout.
#[repr(C)]
#[derive(Clone, Copy, Default)]
pub struct SkillOutcome {
    /// 1 when the skill won the race, -1 when it lost a won race, otherwise 0.
    win_delta: i32,
    /// Change of the team trial score before bonuses.
    score_delta: i32,
}

/// Returns the text of a panic.
fn panic_text(panic: Box<dyn Any + Send>) -> String {
    match (panic.downcast_ref::<String>(), panic.downcast_ref::<&str>()) {
        (Some(text), _) => text.clone(),
        (None, Some(text)) => text.to_string(),
        (None, None) => "the race engine stopped unexpectedly".to_string(),
    }
}

/// Runs a fallible body, keeping a failure or panic from crossing into the caller. The body's value is returned, or
/// `fallback` after storing the failure text.
fn guarded<T>(fallback: T, body: impl FnOnce() -> Result<T, String>) -> T {
    let failure = match catch_unwind(AssertUnwindSafe(body)) {
        Ok(Ok(value)) => return value,
        Ok(Err(text)) => text,
        Err(panic) => panic_text(panic),
    };
    if let Ok(mut last) = LAST_ERROR.lock() {
        *last = failure;
    }
    fallback
}

/// Returns the version of this interface.
#[unsafe(no_mangle)]
pub extern "C" fn racesim_interface_version() -> u32 {
    INTERFACE_VERSION
}

/// Copies the text of the last failure into `buffer` and returns the number of bytes copied.
///
/// # Safety
/// `buffer` must be writable for `capacity` bytes.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn racesim_last_error(buffer: *mut u8, capacity: usize) -> usize {
    let Ok(last) = LAST_ERROR.lock() else { return 0 };
    let length = last.len().min(capacity);
    unsafe { std::ptr::copy_nonoverlapping(last.as_ptr(), buffer, length) };
    length
}

/// Loads the current game data from a folder, given as UTF-8. Returns null on failure. The data lasts for the life of
/// the process.
///
/// # Safety
/// `directory` must be readable for `length` bytes.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn racesim_open(directory: *const u8, length: usize) -> *mut GameData {
    guarded(std::ptr::null_mut(), || {
        let bytes = unsafe { std::slice::from_raw_parts(directory, length) };
        let path = std::str::from_utf8(bytes).map_err(|error| error.to_string())?;
        let data = GameData::load(Path::new(path), DataEra::current())?;
        Ok(Box::into_raw(Box::new(data)))
    })
}

/// Reads the field, course and conditions of a saved race, given as JSON. The recorded result is not used. Returns
/// null on failure. `racesim_release` frees the race.
///
/// # Safety
/// `json` must be readable for `length` bytes.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn racesim_prepare_saved_race(json: *const u8, length: usize) -> *mut RaceInput {
    guarded(std::ptr::null_mut(), || {
        let bytes = unsafe { std::slice::from_raw_parts(json, length) };
        let text = std::str::from_utf8(bytes).map_err(|error| error.to_string())?;
        let (mut input, _) = corpus::parse_race(text)?;
        // The engine does not implement race items
        for horse in &mut input.horses {
            horse.item_ids.clear();
        }
        Ok(Box::into_raw(Box::new(input)))
    })
}

/// Frees a race from `racesim_prepare_saved_race`.
///
/// # Safety
/// `race` must come from `racesim_prepare_saved_race` and must not be used afterward.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn racesim_release(race: *mut RaceInput) {
    drop(unsafe { Box::from_raw(race) });
}

/// Simulates one race per seed on all processor threads and stores the outcome of each runner and every skill span.
/// The outcomes of seed `k` occupy the elements from `k * runners` in gate order, and the skill spans are in seed
/// order, so the result does not depend on the thread count. Returns 0 on success. One caller at a time may use the
/// race. The data is read only.
///
/// # Safety
/// `data` must come from `racesim_open` and `race` from `racesim_prepare_saved_race`. `seeds` must be readable for
/// `seed_count` elements. `output` must be writable, and its `outcomes` writable for `outcome_count` elements.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn racesim_run(
    data: *const GameData,
    race: *const RaceInput,
    seeds: *const i32,
    seed_count: usize,
    output: *mut TrialOutput,
) -> i32 {
    guarded(1, || {
        let (data, race, output) = unsafe { (&*data, &*race, &mut *output) };
        let seeds = unsafe { std::slice::from_raw_parts(seeds, seed_count) };
        let outcome_count = output.outcome_count;
        let outcomes = unsafe { std::slice::from_raw_parts_mut(output.outcomes, outcome_count) };
        let runners = race.horses.len();
        if outcome_count != seed_count * runners {
            return Err(format!("{seed_count} races of {runners} runners cannot fill {outcome_count} outcomes"));
        }
        let threads = thread::available_parallelism().map_or(1, NonZero::get).min(seed_count);
        let seeds_per_thread = seed_count.div_ceil(threads);
        let spans = thread::scope(|scope| {
            let workers: Vec<_> = seeds
                .chunks(seeds_per_thread)
                .zip(outcomes.chunks_mut(seeds_per_thread * runners))
                .enumerate()
                .map(|(worker, (seeds, outcomes))| {
                    scope.spawn(move || {
                        let mut input = race.clone();
                        let mut spans = Vec::new();
                        let races = seeds.iter().zip(outcomes.chunks_mut(runners)).enumerate();
                        for (offset, (&seed, outcomes)) in races {
                            input.random_seed = seed;
                            let race_index = (worker * seeds_per_thread + offset) as i32;
                            run_race(&input, data, race_index, outcomes, &mut spans);
                        }
                        spans
                    })
                })
                .collect();
            let mut spans = Vec::new();
            for worker in workers {
                spans.extend(worker.join().map_err(panic_text)?);
            }
            Ok::<_, String>(spans)
        })?;
        let spans = spans.into_boxed_slice();
        output.skill_span_count = spans.len();
        output.skill_spans = Box::into_raw(spans).cast();
        Ok(0)
    })
}

/// Frees the skill spans of a `TrialOutput`.
///
/// # Safety
/// `spans` and `count` must come from one `TrialOutput`, and the spans must not be used afterward.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn racesim_release_skill_spans(spans: *mut SkillSpan, count: usize) {
    drop(unsafe { Box::from_raw(std::ptr::slice_from_raw_parts_mut(spans, count)) });
}

/// Stores 1 for each skill of a trial that can activate for the runner on the trial's course under any of the trial's
/// variations, otherwise 0. A skill counts as unable to activate only when a fixed fact of the race rules it out, such
/// as the course, the ground or the running style. Returns 0 on success.
///
/// # Safety
/// `data` must come from `racesim_open` and `race` from `racesim_prepare_saved_race`. The arrays of `trial` must be
/// readable for their counts, and `possible` writable for `skill_count` elements.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn racesim_possible_skills(
    data: *const GameData,
    race: *const RaceInput,
    trial: *const SkillTrial,
    possible: *mut u8,
) -> i32 {
    guarded(1, || {
        let (data, base, trial) = unsafe { (&*data, &*race, &*trial) };
        let skills = unsafe { std::slice::from_raw_parts(trial.skill_ids, trial.skill_count) };
        let variations = unsafe { std::slice::from_raw_parts(trial.variations, trial.variation_count) };
        let possible = unsafe { std::slice::from_raw_parts_mut(possible, trial.skill_count) };
        possible.fill(0);
        for variation in variations {
            let (input, gate) = varied_race(base, trial, variation);
            // The check draws from the generator of its race, so that race is not run
            let mut unused_race = race::init(&input, data);
            for (flag, &skill_id) in possible.iter_mut().zip(skills) {
                if *flag == 0 && !skill::never_activates(&mut unused_race, gate, skill_id) {
                    *flag = 1;
                }
            }
        }
        Ok(0)
    })
}

/// Simulates one race per variation with every skill of the trial silently watching, which is the baseline race bit for
/// bit, and stores the baseline of each variation and which skills activated. Leaves one processor thread free for the
/// window. Returns 0 on success. One caller at a time may use the race. The data is read only.
///
/// # Safety
/// `data` must come from `racesim_open` and `race` from `racesim_prepare_saved_race`. The arrays of `trial` must be
/// readable or writable for their counts, and `score_rules` readable.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn racesim_watch_skills(data: *const GameData, race: *const RaceInput, trial: *const SkillTrial) -> i32 {
    guarded(1, || {
        let (data, base, trial) = unsafe { (&*data, &*race, &*trial) };
        let rules = unsafe { &*trial.score_rules };
        let (skill_count, variation_count) = (trial.skill_count, trial.variation_count);
        let skills = unsafe { std::slice::from_raw_parts(trial.skill_ids, skill_count) };
        let replaced = unsafe { std::slice::from_raw_parts(trial.replaced_skill_ids, skill_count) };
        let variations = unsafe { std::slice::from_raw_parts(trial.variations, variation_count) };
        let baselines = unsafe { std::slice::from_raw_parts_mut(trial.baselines, variation_count) };
        let activations = unsafe { std::slice::from_raw_parts_mut(trial.activations, skill_count * variation_count) };
        // A trial without skills has no activation rows and still has baselines
        let rows = activations.chunks_mut(skill_count.max(1)).chain(std::iter::repeat_with(|| &mut [][..]));
        let races: Vec<_> = variations.iter().map(|variation| varied_race(base, trial, variation)).collect();
        for_each_on_threads(
            races.into_iter().zip(baselines).zip(rows),
            || (),
            |_, (((mut input, gate), baseline), row)| {
                let horse = &mut input.horses[gate];
                horse.skills.extend(skills.iter().map(|&skill_id| SkillInput { skill_id, level: 1 }));
                horse.muted_skills = skills.to_vec();
                horse.observed_skills = skills.to_vec();
                let (race, positioning_score) = simulate_to_finish(&input, data, gate, rules);
                row.fill(0);
                for span in race.recorder.skill_spans.iter().filter(|span| span.horse == gate) {
                    if let Ok(k) = skills.binary_search(&span.skill_id) {
                        row[k] |= FIRED_SKILL;
                    }
                    for (flag, _) in row.iter_mut().zip(replaced).filter(|(_, tier)| **tier == span.skill_id) {
                        *flag |= FIRED_REPLACED_TIER;
                    }
                }
                *baseline = BaselineOutcome {
                    score: positioning_score + race_score(&race, &input, gate, rules),
                    finish_order: race.recorder.results[gate].finish_order,
                };
            },
        )?;
        Ok(0)
    })
}

/// For each skill of the trial and each variation stored by `racesim_watch_skills` in which the skill or the tier that
/// it replaces activated, simulates the race with that skill in effect and stores what it changed against the baseline.
/// A race in which neither activated is the baseline race, so it is not repeated and stores no change. Each skill race
/// continues from a stop of the race shared by the skills of its variation, where its skill allows that. The outcomes of
/// skill `k` occupy the elements from `k * variation_count` in variation order. Leaves one processor thread free for
/// the window. Returns 0 on success. One caller at a time may use the race. The data is read only.
///
/// # Safety
/// `data` must come from `racesim_open` and `race` from `racesim_prepare_saved_race`. The arrays of `trial` must be
/// readable for their counts and `score_rules` readable. `outcomes` must be writable for
/// `skill_count * variation_count` elements.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn racesim_run_skill_trial(
    data: *const GameData,
    race: *const RaceInput,
    trial: *const SkillTrial,
    outcomes: *mut SkillOutcome,
) -> i32 {
    guarded(1, || {
        let (data, base, trial) = unsafe { (&*data, &*race, &*trial) };
        let rules = unsafe { &*trial.score_rules };
        let (skill_count, variation_count) = (trial.skill_count, trial.variation_count);
        let skills = unsafe { std::slice::from_raw_parts(trial.skill_ids, skill_count) };
        let replaced = unsafe { std::slice::from_raw_parts(trial.replaced_skill_ids, skill_count) };
        let variations = unsafe { std::slice::from_raw_parts(trial.variations, variation_count) };
        let baselines = unsafe { std::slice::from_raw_parts(trial.baselines, variation_count) };
        let activations = unsafe { std::slice::from_raw_parts(trial.activations, skill_count * variation_count) };
        let outcomes = unsafe { std::slice::from_raw_parts_mut(outcomes, skill_count * variation_count) };
        let races: Vec<_> = variations.iter().map(|variation| varied_race(base, trial, variation)).collect();
        let can_share: Vec<_> = skills.iter().map(|&id| skill::stays_baseline_until_first_effect(data, base, trial.runner, id)).collect();
        let mut shared_races: Vec<_> = races.iter().map(|_| None).collect();
        for_each_on_threads(
            shared_races.iter_mut().enumerate(),
            || (),
            |_, (variation, shared)| {
                let row = &activations[variation * skill_count..][..skill_count];
                let (kept, tiers): (Vec<_>, Vec<_>) =
                    (0..skill_count).filter(|&k| row[k] != 0 && can_share[k]).map(|k| (skills[k], replaced[k])).unzip();
                if !kept.is_empty() {
                    let (input, gate) = &races[variation];
                    *shared = Some(SharedRace::new(input, *gate, data, [&kept, &tiers], rules));
                }
            },
        )?;
        for_each_on_threads(
            outcomes.iter_mut().enumerate(),
            || races.clone(),
            |races, (position, outcome)| {
                let (skill, variation) = (position / variation_count, position % variation_count);
                *outcome = SkillOutcome::default();
                if activations[variation * skill_count + skill] != 0 {
                    let (input, gate) = &mut races[variation];
                    let stop = shared_races[variation].as_ref().and_then(|shared: &SharedRace| shared.stop_before(skills[skill]));
                    *outcome = run_skill_race(input, *gate, data, [skills[skill], replaced[skill]], stop, &baselines[variation], rules);
                }
            },
        )?;
        Ok(0)
    })
}

/// Returns the race of a variation on the trial's course and the gate of the runner under test.
fn varied_race(base: &RaceInput, trial: &SkillTrial, variation: &RaceVariation) -> (RaceInput, usize) {
    let order = &variation.gate_order[..base.horses.len()];
    let mut horses: Vec<_> = order.iter().map(|&index| base.horses[index as usize].clone()).collect();
    for (gate, horse) in horses.iter_mut().enumerate() {
        horse.frame_order = gate as i32 + 1;
    }
    let gate = order.iter().position(|&index| index as usize == trial.runner);
    let input = RaceInput {
        random_seed: variation.seed,
        course_set_id: trial.course_set_id,
        ground_condition: variation.ground_condition,
        weather: variation.weather,
        season: variation.season,
        time: base.time,
        race_type: base.race_type,
        horses,
        race_instance_id: trial.race_instance_id,
    };
    (input, gate.expect("every variation holds the runner under test at a gate"))
}

/// A variation of a trial up to the first effects of its skills, from which each skill race continues.
struct SharedRace<'a> {
    /// In race order, none past the first random skill pick of the runner under test, because an observed skill can
    /// take the place of a pick.
    stops: Vec<RaceStop<'a>>,
    /// The skills that continue from a stop, in ascending order.
    skills: Vec<i32>,
    /// For each skill, the race time of the first effect of the skill or of the tier that it replaces.
    first_effects: Vec<f32>,
}

/// One point of a shared race, between two steps.
struct RaceStop<'a> {
    /// The race without any skill under test.
    race: Race<'a>,
    progress: PositioningProgress,
    /// What the runner under test holds at this point in the race that watches the skills.
    watchers: SkillManager,
}

impl<'a> SharedRace<'a> {
    /// Runs the race that watches the skills, which finds their first effects and holds their state at each stop, then
    /// the race without them up to the last stop that any skill can use. Takes the skills and their replaced tiers.
    fn new(input: &'a RaceInput, gate: usize, data: &'a GameData, [skills, tiers]: [&[i32]; 2], rules: &ScoreRules) -> Self {
        let mut watch_input = input.clone();
        let horse = &mut watch_input.horses[gate];
        horse.skills.extend(skills.iter().map(|&skill_id| SkillInput { skill_id, level: 1 }));
        horse.muted_skills = skills.to_vec();
        horse.observed_skills = skills.to_vec();
        let mut watch = started_race(&watch_input, data);
        let mut watchers = Vec::new();
        let mut steps = 0;
        let mut finished = false;
        let mut picked = false;
        while !finished {
            picked |= !watch.horses[gate].skill_manager.reserved_skill_ids.is_empty();
            if steps % STOP_INTERVAL_STEPS == 0 && !picked {
                watchers.push(watch.horses[gate].skill_manager.clone());
            }
            finished = tick::step(&mut watch);
            steps += 1;
        }
        let mut first_effects = vec![f32::INFINITY; skills.len()];
        for span in watch.recorder.skill_spans.iter().filter(|span| span.horse == gate) {
            for (k, first_effect) in first_effects.iter_mut().enumerate() {
                if skills[k] == span.skill_id || tiers[k] == span.skill_id {
                    *first_effect = first_effect.min(span.start);
                }
            }
        }
        let latest = first_effects.iter().copied().fold(0.0, f32::max);
        let mut race = started_race(input, data);
        let mut progress = PositioningProgress::default();
        let mut stops = Vec::with_capacity(watchers.len());
        let mut watchers = watchers.into_iter().peekable();
        while let Some(held) = watchers.next() {
            if !(race.accumulate_time_since_start < latest) {
                break;
            }
            stops.push(RaceStop { race: race.clone(), progress, watchers: held });
            if watchers.peek().is_some() {
                for _ in 0..STOP_INTERVAL_STEPS {
                    step_scored(&mut race, input, gate, rules, &mut progress);
                }
            }
        }
        SharedRace { stops, skills: skills.to_vec(), first_effects }
    }

    /// Returns the last stop before the first effect of a skill, if the skill continues from a stop.
    fn stop_before(&self, skill_id: i32) -> Option<&RaceStop<'a>> {
        let first_effect = self.first_effects[self.skills.binary_search(&skill_id).ok()?];
        self.stops.iter().rev().find(|stop| stop.race.accumulate_time_since_start < first_effect)
    }
}

/// Adds a skill to the runner at `gate` in effect, simulates the race, removes the skill again and returns what it
/// changed. The ids are the skill and the tier that it replaces, or 0. The skill draws from its own generator, so the
/// race is the baseline race up to its first effect and the difference measures that skill alone, which also lets the
/// race continue from a stop before that effect. An upgrade mutes the tier that it replaces, because the game holds only
/// one tier of a skill.
fn run_skill_race(
    input: &mut RaceInput,
    gate: usize,
    data: &GameData,
    [skill_id, replaced_id]: [i32; 2],
    stop: Option<&RaceStop>,
    baseline: &BaselineOutcome,
    rules: &ScoreRules,
) -> SkillOutcome {
    let horse = &mut input.horses[gate];
    horse.skills.push(SkillInput { skill_id, level: 1 });
    horse.observed_skills.push(skill_id);
    horse.muted_skills.push(replaced_id);
    let (race, positioning_score) = match stop {
        Some(stop) => {
            let mut race = stop.race.clone();
            race.input = input;
            let watched = skill::find_skill(&stop.watchers, skill_id).expect("a stop holds the skills that continue from it");
            race.horses[gate].skill_manager.adopt_observed_skill(&stop.watchers, watched);
            finish_race(race, input, gate, rules, stop.progress)
        }
        None => finish_race(started_race(input, data), input, gate, rules, PositioningProgress::default()),
    };
    let (score, finish_order) = (positioning_score + race_score(&race, input, gate, rules), race.recorder.results[gate].finish_order);
    drop(race);
    let horse = &mut input.horses[gate];
    horse.skills.pop();
    horse.observed_skills.clear();
    horse.muted_skills.clear();
    SkillOutcome {
        win_delta: (finish_order == 0) as i32 - (baseline.finish_order == 0) as i32,
        score_delta: score - baseline.score,
    }
}

/// What the score of the runner under test has gathered so far for being ahead going into each phase.
#[derive(Clone, Copy, Default)]
struct PositioningProgress {
    phases_entered: usize,
    score: i32,
}

/// Returns a race between its start and its first step, recording no frames, which a score does not read.
fn started_race<'a>(input: &'a RaceInput, data: &'a GameData) -> Race<'a> {
    let mut race = race::init(input, data);
    race.record_frames = false;
    tick::start(&mut race);
    race
}

/// Runs one step and adds up the positioning score of the runner at `gate`. Returns whether every runner has finished.
fn step_scored(race: &mut Race, input: &RaceInput, gate: usize, rules: &ScoreRules, progress: &mut PositioningProgress) -> bool {
    // The game judges a phase change by the distances before the step
    let ahead = is_ahead_of_same_style(race, input, gate);
    let finished = tick::step(race);
    while progress.phases_entered < rules.positioning.len() && race.horses[gate].phase as usize > progress.phases_entered {
        if ahead {
            progress.score += rules.positioning[progress.phases_entered];
        }
        progress.phases_entered += 1;
    }
    finished
}

/// Runs a race to its finish and returns it with the score of the runner at `gate` for being ahead going into each phase.
fn finish_race<'a>(
    mut race: Race<'a>,
    input: &RaceInput,
    gate: usize,
    rules: &ScoreRules,
    mut progress: PositioningProgress,
) -> (Race<'a>, i32) {
    while !step_scored(&mut race, input, gate, rules, &mut progress) {}
    tick::finish(&mut race);
    (race, progress.score)
}

/// Simulates one race and returns the finished race with the score of the runner at `gate` for being ahead going into
/// each phase.
fn simulate_to_finish<'a>(input: &'a RaceInput, data: &'a GameData, gate: usize, rules: &ScoreRules) -> (Race<'a>, i32) {
    finish_race(started_race(input, data), input, gate, rules, PositioningProgress::default())
}

/// Returns whether the runner at `gate` is ahead of all other runners of its running style.
fn is_ahead_of_same_style(race: &Race, input: &RaceInput, gate: usize) -> bool {
    let (style, distance) = (input.horses[gate].running_style, race.horses[gate].distance);
    let runners = input.horses.iter().zip(&race.horses).enumerate();
    runners.filter(|(other, (entry, _))| *other != gate && entry.running_style == style).all(|(_, (_, h))| h.distance < distance)
}

/// Returns the team trial score before bonuses of the runner at `gate` in a finished race, without the score for being
/// ahead going into each phase. The activations of a muted skill do not score.
fn race_score(race: &Race, input: &RaceInput, gate: usize, rules: &ScoreRules) -> i32 {
    let results = &race.recorder.results;
    let result = &results[gate];
    let entry = &input.horses[gate];
    let mut score = rules.place.get(result.finish_order as usize).copied().unwrap_or(0);
    if result.finish_order == 0 {
        let lengths = results.iter().find(|r| r.finish_order == 1).map_or(0.0, |r| r.finish_diff_time * LENGTHS_PER_SECOND);
        let margin = MARGIN_LENGTH_LIMITS.iter().position(|&limit| lengths < limit).unwrap_or(MARGIN_LENGTH_LIMITS.len());
        score += rules.margin[margin];
    }
    if result.finish_order < rules.long_shot_place && entry.popularity >= rules.long_shot_popularity {
        score += rules.long_shot;
    }
    // The game compares with the standard course time rounded up to a tenth of a second
    let standard_tenths = (race.info.border_time_scaled as f64 * 10.0).ceil();
    let tenths_under = (standard_tenths - result.finish_time_scaled as f64 * 10.0).floor() as i32;
    if tenths_under > 0 {
        score += rules.target_time[tenths_under.min(LAST_TARGET_TIME_BRACKET) as usize];
    }
    let spans = race.recorder.spans.iter().filter(|span| span.horse == gate && span.kind == span_kind::RUSHED);
    let rushed_seconds: f32 = spans.map(|span| span.end - span.start).sum();
    if rushed_seconds > 0.0 {
        // The first second of a rush does not count
        score += rules.rushed + rules.rushed_per_second * (rushed_seconds as i32 - 1).max(0);
    }
    if result.start_delay_time < STRONG_START_DELAY {
        score += rules.strong_start;
    }
    score + skill_score(race, entry, gate, rules)
}

/// Returns the score for the skill activations of the runner at `gate`.
fn skill_score(race: &Race, entry: &HorseInput, gate: usize, rules: &ScoreRules) -> i32 {
    let events = race.recorder.events.iter().filter(|event| event.kind == event_kind::SKILL && event.params[0] == gate as i32);
    events
        .map(|event| event.params[1])
        .filter(|skill_id| !entry.muted_skills.contains(skill_id))
        .map(|skill_id| match race.data.skills.get(&skill_id).map_or(0, |skill| skill.rarity) {
            RARITY_NORMAL => rules.normal_skill,
            RARITY_RARE => rules.rare_skill,
            rarity if rarity >= RARITY_FIRST_UNIQUE => {
                let level = entry.skills.iter().find(|skill| skill.skill_id == skill_id).map_or(1, |skill| skill.level);
                rules.unique_skill[(rarity - RARITY_FIRST_UNIQUE) as usize][(level - 1) as usize]
            }
            _ => 0,
        })
        .sum()
}

/// Applies `work` to each item on all processor threads but one, each thread with its own state from `state`.
fn for_each_on_threads<T: Send, S>(
    items: impl Iterator<Item = T> + Send,
    state: impl Fn() -> S + Sync,
    work: impl Fn(&mut S, T) + Sync,
) -> Result<(), String> {
    let queue = Mutex::new(items);
    let threads = thread::available_parallelism().map_or(1, NonZero::get).saturating_sub(1).max(1);
    thread::scope(|scope| {
        let workers: Vec<_> = (0..threads)
            .map(|_| {
                scope.spawn(|| {
                    let mut thread_state = state();
                    loop {
                        let next = queue.lock().unwrap_or_else(|poisoned| poisoned.into_inner()).next();
                        let Some(item) = next else { break };
                        work(&mut thread_state, item);
                    }
                })
            })
            .collect();
        for worker in workers {
            worker.join().map_err(panic_text)?;
        }
        Ok(())
    })
}

/// Simulates one race, stores the outcome of each runner in gate order and adds its skill spans to `spans`.
fn run_race(
    input: &RaceInput,
    data: &GameData,
    race_index: i32,
    outcomes: &mut [RunnerOutcome],
    spans: &mut Vec<SkillSpan>,
) {
    outcomes.fill(RunnerOutcome::default());
    let mut race = race::init(input, data);
    tick::start(&mut race);
    loop {
        let all_finished = tick::step(&mut race);
        observe_step(&race, outcomes);
        if all_finished {
            break;
        }
    }
    tick::finish(&mut race);
    store_result(&race, outcomes);
    store_mode_times(&race, outcomes);
    store_finish_gaps(&race, outcomes);
    store_skill_spans(&race, race_index, spans);
}

/// Adds the step to the values that the race result does not hold. A runner counts from the opening of its gate
/// until the step in which it finishes.
fn observe_step(race: &Race, outcomes: &mut [RunnerOutcome]) {
    let goal = race.info.race_course_set.distance as f32;
    let time = race.accumulate_time_since_start;
    for (runner, (h, outcome)) in race.horses.iter().zip(outcomes).enumerate() {
        if h.is_finished || h.delay_time > 0.0 {
            continue;
        }
        // The speed covers the part of the step after the gate opens. The move falls short of it only through the
        // limit of the world transform.
        let planned_move = h.last_speed * (time - h.delay_time_saved).min(DT);
        outcome.world_transform_loss += planned_move - (h.distance - h.prev_distance);
        if h.ai.is_down_slope_accel_mode {
            if h.phase >= Phase::End {
                outcome.downhill_time_late += DT;
            } else {
                outcome.downhill_time_early += DT;
            }
        }
        if h.hp > 0.0 {
            continue;
        }
        if outcome.hp_out_distance == 0.0 && goal - h.distance > HP_OUT_GOAL_MARGIN {
            outcome.hp_out_distance = goal - h.distance;
        }
        let guts = stats::guts(h);
        let per_second = hp::calc_dec_hp_per_sec(
            race,
            runner,
            race.info.base_speed,
            h.ground_modifier_param.multi_hp_sub,
            h.phase,
            guts,
            h.last_self_speed,
        );
        outcome.hp_shortfall += per_second * DT;
    }
}

/// Stores the values that the end state of the race holds.
fn store_result(race: &Race, outcomes: &mut [RunnerOutcome]) {
    let late_race = race.info.phase_distances.phase_end_start_distance;
    let results = race.horses.iter().zip(&race.recorder.results);
    for ((h, result), outcome) in results.zip(outcomes) {
        outcome.finish_order = result.finish_order;
        outcome.finish_time = result.finish_time_raw;
        outcome.start_delay = result.start_delay_time;
        outcome.late_start = (result.start_delay_time > DT) as i32;
        let spurt_start = result.last_spurt_start_distance;
        outcome.spurt_delay = if spurt_start < 0.0 { -1.0 } else { (spurt_start - late_race).max(0.0) };
        outcome.spurt_speed = h.ai.last_spurt_target_speed;
        outcome.start_hp = h.max_hp;
        outcome.final_hp = h.hp;
    }
}

/// Adds up the time that each runner spent in the modes that the race records as spans.
fn store_mode_times(race: &Race, outcomes: &mut [RunnerOutcome]) {
    for span in &race.recorder.spans {
        let outcome = &mut outcomes[span.horse];
        let length = span.end - span.start;
        match span.kind {
            span_kind::DUEL => outcome.duel_time += length,
            span_kind::SPEED_UP | span_kind::OVERTAKE | span_kind::PACE_UP | span_kind::PACE_UP_EX => {
                outcome.pace_up_time += length;
            }
            span_kind::PACE_DOWN => outcome.pace_down_time += length,
            _ => {}
        }
    }
}

/// Stores how far short of the goal each runner was when the previous finisher finished.
fn store_finish_gaps(race: &Race, outcomes: &mut [RunnerOutcome]) {
    let goal = race.info.race_course_set.distance as f32;
    let results = &race.recorder.results;
    let mut finishers: Vec<usize> = (0..results.len()).collect();
    finishers.sort_unstable_by_key(|&runner| results[runner].finish_order);
    for pair in finishers.windows(2) {
        let distance = distance_at(&race.recorder.frames, pair[1], results[pair[0]].finish_time_raw);
        outcomes[pair[1]].finish_gap = (goal - distance).max(0.0);
    }
}

/// Adds the skill spans of the race, placed on the course. A skill still active at the end of the race ends at the
/// goal.
fn store_skill_spans(race: &Race, race_index: i32, spans: &mut Vec<SkillSpan>) {
    let goal = race.info.race_course_set.distance as f32;
    let frames = &race.recorder.frames;
    spans.extend(race.recorder.skill_spans.iter().map(|span| SkillSpan {
        race: race_index,
        runner: span.horse as i32,
        skill_id: span.skill_id,
        passive: span.infinite as i32,
        start_distance: distance_at(frames, span.horse, span.start).min(goal),
        end_distance: span.end.map_or(goal, |end| distance_at(frames, span.horse, end).min(goal)),
    }));
}

/// Returns the distance of a runner at a time, interpolated between the recorded frames around that time.
fn distance_at(frames: &[Frame], runner: usize, time: f32) -> f32 {
    let next = frames.partition_point(|frame| frame.time < time);
    if next == 0 || next == frames.len() {
        return frames[next.min(frames.len() - 1)].horses[runner].distance;
    }
    let (before, after) = (&frames[next - 1], &frames[next]);
    let share = (time - before.time) / (after.time - before.time);
    let (from, to) = (before.horses[runner].distance, after.horses[runner].distance);
    from + (to - from) * share
}
