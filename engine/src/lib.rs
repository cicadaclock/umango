//! C interface of the race engine, which the Umango application loads at run time.

use std::panic::{AssertUnwindSafe, catch_unwind};
use std::path::Path;
use std::sync::Mutex;

use uma_race_sim::api::RaceInput;
use uma_race_sim::data::{DataEra, GameData, corpus};
use uma_race_sim::sim;

/// Version of this interface, checked by the application when it loads the library.
const INTERFACE_VERSION: u32 = 1;

/// Text of the last failure, read through `racesim_last_error`.
static LAST_ERROR: Mutex<String> = Mutex::new(String::new());

/// Runs a fallible body, keeping a failure or panic from crossing into the caller. The body's value is returned, or
/// `fallback` after storing the failure text.
fn guarded<T>(fallback: T, body: impl FnOnce() -> Result<T, String>) -> T {
    let failure = match catch_unwind(AssertUnwindSafe(body)) {
        Ok(Ok(value)) => return value,
        Ok(Err(text)) => text,
        Err(panic) => match (panic.downcast_ref::<String>(), panic.downcast_ref::<&str>()) {
            (Some(text), _) => text.clone(),
            (None, Some(text)) => text.to_string(),
            (None, None) => "the race engine stopped unexpectedly".to_string(),
        },
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

/// Simulates one race from a seed and stores the finish order of each runner, counted from 0, in gate order. Returns
/// 0 on success. The race and data may be used from one thread at a time.
///
/// # Safety
/// `data` must come from `racesim_open`, `race` from `racesim_prepare_saved_race`, and `finish_order` must be writable
/// for `count` elements.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn racesim_run(
    data: *const GameData,
    race: *mut RaceInput,
    seed: i32,
    finish_order: *mut i32,
    count: usize,
) -> i32 {
    guarded(1, || {
        let (data, race) = unsafe { (&*data, &mut *race) };
        race.random_seed = seed;
        let output = sim::simulate(race, data);
        if output.results.len() != count {
            return Err(format!("the race has {} runners, expected {count}", output.results.len()));
        }
        let orders = unsafe { std::slice::from_raw_parts_mut(finish_order, count) };
        for (order, result) in orders.iter_mut().zip(&output.results) {
            if !(0..count as i32).contains(&result.finish_order) {
                return Err(format!("the race holds finish order {} of {count} runners", result.finish_order));
            }
            *order = result.finish_order;
        }
        Ok(0)
    })
}
