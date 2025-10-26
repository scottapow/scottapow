let formatter = Intl.DateTimeFormat(undefined, {
  hour12: false,
  hourCycle: 'h24',
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
});

export function getInputDate(d: Date) {
  let parts = getParts(d);
  return `${parts.year}-${parts.month}-${parts.day}`;
}

export function getInputTime(d: Date) {
  let parts = getParts(d);
  return `${parts.hour}:${parts.minute}`;
}

const getParts = (d: Date) => formatter.formatToParts(d).reduce<Record<Intl.DateTimeFormatPartTypes, string>>(
  (acc, { type, value }) => {
    acc[type] = value;
    return acc;
  }, {} as Record<Intl.DateTimeFormatPartTypes, string>
);