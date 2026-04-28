import { describe, expect, it } from 'vitest'
import { convertRichHtmlToMarkdown, getClipboardImageFiles } from './richPaste'

describe('convertRichHtmlToMarkdown', () => {
  it('converts common rich text clipboard html to markdown', () => {
    const markdown = convertRichHtmlToMarkdown(`
      <h2>Launch notes</h2>
      <p>Hello <strong>Decoy</strong>, read <a href="https://example.com/path?q=1">the docs</a>.</p>
      <ul><li>First item</li><li><em>Second</em> item</li></ul>
    `)

    expect(markdown).toBe([
      '## Launch notes',
      '',
      'Hello **Decoy**, read [the docs](https://example.com/path?q=1).',
      '',
      '- First item',
      '- *Second* item',
    ].join('\n'))
  })
})

describe('getClipboardImageFiles', () => {
  it('extracts image files from clipboard files and items', () => {
    const directImage = { name: 'direct.png', type: 'image/png' } as File
    const textFile = { name: 'note.txt', type: 'text/plain' } as File
    const itemImage = { name: 'item.jpg', type: 'image/jpeg' } as File

    const files = getClipboardImageFiles({
      files: [directImage, textFile],
      items: [
        { kind: 'file', type: 'image/jpeg', getAsFile: () => itemImage },
        { kind: 'string', type: 'text/html', getAsFile: () => null },
      ],
    })

    expect(files).toEqual([directImage, itemImage])
  })
})
