import { describe, expect, it } from 'vitest'
import { renderMarkdown } from './markdown'

describe('renderMarkdown', () => {
  it('renders rich paste markdown into html', () => {
    expect(renderMarkdown([
      '## Launch notes',
      '',
      'Hello **Decoy**, read [the docs](https://example.com/path?q=1).',
      '',
      '- First item',
      '- *Second* item',
    ].join('\n'))).toContain('<strong>Decoy</strong>')
  })

  it('escapes raw html from markdown content', () => {
    const html = renderMarkdown('<img src=x onerror=alert(1)> **safe**')

    expect(html).toContain('&lt;img')
    expect(html).not.toContain('onerror=')
    expect(html).toContain('<strong>safe</strong>')
  })

  it('does not render markdown images that are already handled by embedded image attachments', () => {
    const html = renderMarkdown('![uploaded](https://cdn.test/upload.png)')

    expect(html).not.toContain('<img')
    expect(html).toContain('uploaded')
  })
})
