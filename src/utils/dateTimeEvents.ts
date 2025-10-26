export const minuteChangeEventType = 'minuteChange' as const;
export const dayChangeEventType = 'dayChange' as const;

const MS_SECOND = 1_000;
const MS_MINUTE = 60 * MS_SECOND;
const MS_HOUR = 60 * MS_MINUTE;
const MS_DAY = 24 * MS_HOUR;

export function watchTime() {
  let now = new Date();
  let msToNextSecond = 999 - now.getMilliseconds();
  let msToNextMinute = (59 - now.getSeconds()) * MS_SECOND;
  let interval: NodeJS.Timeout;

  setTimeout(
    () => triggerMinuteChange(() => {
      interval = setInterval(triggerMinuteChange, MS_MINUTE);
    }),
    msToNextMinute + msToNextSecond
  );

  let msToNextHour = (59 - now.getMinutes()) * MS_MINUTE;
  let msToNextDay = (23 - now.getHours()) * MS_HOUR;

  setTimeout(
    () => triggerDayChange(() => {
      setInterval(triggerDayChange, MS_DAY);
    }),
    msToNextDay + msToNextHour + msToNextMinute + msToNextSecond
  );
}

const triggerMinuteChange = makeTriggerDateTimeChange(minuteChangeEventType);
const triggerDayChange = makeTriggerDateTimeChange(dayChangeEventType);

function makeTriggerDateTimeChange(eventName: string) {
  return function triggerDayChange(cb?: (n: Date) => void) {
    let newNow = new Date();
    if (typeof document !== 'undefined') {
      document.dispatchEvent(new CustomEvent(
        eventName,
        { detail: newNow, bubbles: true }
      ));
      if (typeof cb === 'function') cb(newNow);
    }
  }
}