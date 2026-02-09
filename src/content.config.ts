import { defineCollection, reference } from 'astro:content';
import { glob } from 'astro/loaders';
import { z } from 'astro/zod';

const authors = defineCollection({
  loader: glob({
    pattern: ["**/*.md", "!draft-*"],
    base: './src/content/authors',
  }),
  schema: z.object({
    firstname: z.string(),
    surname: z.string(),
    headshot_url: z.string().optional(),
    description: z.string(),
  })
})

const stories = defineCollection({
  loader: glob({
    pattern: ["**/*.md", "!draft-*"],
    base: './src/content/stories',
  }),
  schema: z.object({
    author: z.array(reference('authors')),
    publish_date: z.coerce.date(),
    slug: z.string(),
  })
});

export const collections = { stories, authors };