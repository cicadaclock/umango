//! C interface of the race engine, which the Umango application loads at run time.

use std::any::Any;
use std::num::NonZero;
use std::panic::{AssertUnwindSafe, catch_unwind};
use std::path::Path;
use std::sync::Mutex;
use std::thread;

use uma_race_sim::api::{Frame, RaceInput, span_kind};
use uma_race_sim::data::{DataEra, GameData, corpus};
use uma_race_sim::sim::defs::Phase;
use uma_race_sim::sim::physics::{hp, stats};
use uma_race_sim::sim::race::{self, DT, Race};
use uma_race_sim::sim::tick;

/// Version of this interface, checked by the application when it loads the library.
const INTERFACE_VERSION: u32 = 3;

/// HP that runs out within this many meters of the goal does not count as running out short of the goal.
const HP_OUT_GOAL_MARGIN: f32 = 0.1;

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
/// race and data.
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
