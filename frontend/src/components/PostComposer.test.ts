import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./PostComposer.vue', import.meta.url), 'utf8')

describe('PostComposer paste handling', () => {
  it('extracts clipboard image files before rich text paste conversion', () => {
    expect(source).toContain('getClipboardImageFiles(event.clipboardData)')
    expect(source.indexOf('getClipboardImageFiles(event.clipboardData)')).toBeLessThan(source.indexOf('convertRichHtmlToMarkdown(html)'))
  })
})
