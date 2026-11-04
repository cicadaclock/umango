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
use uma_race_sim::sim::{skill, tick};

/// Version of this interface, checked by the application when it loads the library.
const INTERFACE_VERSION: u32 = 5;

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
    skill_ids: *const i32,
    /// For each skill, the lower tier that the runner owns and the skill would replace, otherwise 0.
    replaced_skill_ids: *const i32,
    skill_count: usize,
    variations: *const RaceVariation,
    variation_count: usize,
    score_rules: *const ScoreRules,
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

/// What one added skill changed in one race, against the same race with that skill muted. The application holds the
/// same layout.
#[repr(C)]
#[derive(Clone, Copy, Default)]
pub struct SkillPairOutcome {
    /// 1 when the skill activated, otherwise 0.
    activated: i32,
    /// 1 when the skill won the race, -1 when it lost a won race, otherwise 0.
    win_delta: i32,
    /// Change of the team trial score before bonuses.
    score_delta: i32,
}

/// What the runner under test did in one simulated race.
struct RunnerSample {
    score: i32,
    finish_order: i32,
    /// Whether each watched skill activated.
    activated: [bool; 2],
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

/// Simulates the baseline of each variation and, for each skill of the trial and each variation, the race with that
/// skill muted and, when the skill or the tier that it replaces activated, the same race with the skill's effects. A
/// race in which neither activated is the same either way, so it is not repeated. The outcomes of skill `k` occupy the
/// elements from `k * variation_count` in variation order. Leaves one processor thread free for the window. Returns 0
/// on success. One caller at a time may use the race. The data is read only.
///
/// # Safety
/// `data` must come from `racesim_open` and `race` from `racesim_prepare_saved_race`. The arrays of `trial` must be
/// readable for their counts. `baselines` must be writable for `variation_count` elements and `outcomes` for
/// `skill_count * variation_count` elements. `score_rules` of `trial` must be readable.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn racesim_run_skill_trial(
    data: *const GameData,
    race: *const RaceInput,
    trial: *const SkillTrial,
    baselines: *mut BaselineOutcome,
    outcomes: *mut SkillPairOutcome,
) -> i32 {
    guarded(1, || {
        let (data, base, trial) = unsafe { (&*data, &*race, &*trial) };
        let rules = unsafe { &*trial.score_rules };
        let (skill_count, variation_count) = (trial.skill_count, trial.variation_count);
        let skills = unsafe { std::slice::from_raw_parts(trial.skill_ids, skill_count) };
        let replaced = unsafe { std::slice::from_raw_parts(trial.replaced_skill_ids, skill_count) };
        let variations = unsafe { std::slice::from_raw_parts(trial.variations, variation_count) };
        let baselines = unsafe { std::slice::from_raw_parts_mut(baselines, variation_count) };
        let outcomes = unsafe { std::slice::from_raw_parts_mut(outcomes, skill_count * variation_count) };
        let races: Vec<_> = variations.iter().map(|variation| varied_race(base, trial, variation)).collect();
        for_each_on_threads(
            baselines.iter_mut().zip(&races),
            || (),
            |_, (baseline, (input, gate))| {
                let sample = sample_race(input, data, *gate, [0, 0], rules);
                *baseline = BaselineOutcome { score: sample.score, finish_order: sample.finish_order };
            },
        )?;
        for_each_on_threads(
            outcomes.chunks_mut(variation_count).zip(skills.iter().zip(replaced)),
            || races.clone(),
            |races, (outcomes, (&skill_id, &replaced_id))| {
                for (outcome, (input, gate)) in outcomes.iter_mut().zip(races) {
                    *outcome = run_skill_pair(input, *gate, data, [skill_id, replaced_id], rules);
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

/// Adds a skill to the runner at `gate`, simulates the pair of races and removes the skill again. The ids are the skill
/// and the tier that it replaces, or 0. The skill's muted and real races share one generator sequence up to the first
/// effect, so the difference measures that skill alone. An upgrade mutes the tier that it replaces in the real race,
/// because the game holds only one tier of a skill.
fn run_skill_pair(input: &mut RaceInput, gate: usize, data: &GameData, ids: [i32; 2], rules: &ScoreRules) -> SkillPairOutcome {
    let [skill_id, replaced_id] = ids;
    let horse = &mut input.horses[gate];
    horse.skills.push(SkillInput { skill_id, level: 1 });
    horse.muted_skills.push(skill_id);
    let muted = sample_race(input, data, gate, ids, rules);
    let mut outcome = SkillPairOutcome { activated: muted.activated[0] as i32, ..Default::default() };
    if muted.activated[0] || muted.activated[1] {
        let horse = &mut input.horses[gate];
        horse.muted_skills.clear();
        horse.muted_skills.push(replaced_id);
        let real = sample_race(input, data, gate, [0, 0], rules);
        outcome.win_delta = (real.finish_order == 0) as i32 - (muted.finish_order == 0) as i32;
        outcome.score_delta = real.score - muted.score;
    }
    let horse = &mut input.horses[gate];
    horse.skills.pop();
    horse.muted_skills.clear();
    outcome
}

/// Simulates one race and returns what the runner at `gate` did and whether each watched skill activated.
fn sample_race(input: &RaceInput, data: &GameData, gate: usize, watched: [i32; 2], rules: &ScoreRules) -> RunnerSample {
    let mut race = race::init(input, data);
    tick::start(&mut race);
    let (mut phases_entered, mut positioning_score) = (0, 0);
    loop {
        // The game judges a phase change by the distances before the step
        let ahead = is_ahead_of_same_style(&race, input, gate);
        let finished = tick::step(&mut race);
        while phases_entered < rules.positioning.len() && race.horses[gate].phase as usize > phases_entered {
            if ahead {
                positioning_score += rules.positioning[phases_entered];
            }
            phases_entered += 1;
        }
        if finished {
            break;
        }
    }
    tick::finish(&mut race);
    let activated = |skill_id| race.recorder.skill_spans.iter().any(|span| span.horse == gate && span.skill_id == skill_id);
    RunnerSample {
        score: positioning_score + race_score(&race, input, gate, rules),
        finish_order: race.recorder.results[gate].finish_order,
        activated: watched.map(activated),
    }
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
