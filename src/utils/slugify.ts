export default function slugify(text: string): string {
  return text.replaceAll(/\s/g, '-').toLocaleLowerCase();
}