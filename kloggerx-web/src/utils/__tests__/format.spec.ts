import { describe, it, expect } from 'vitest'
import { formatFileSize, formatDate, getDocTypeIcon } from '../format'

describe('formatFileSize', () => {
  it('returns "0 B" for 0 bytes', () => {
    expect(formatFileSize(0)).toBe('0 B')
  })

  it('formats bytes correctly', () => {
    expect(formatFileSize(512)).toBe('512 B')
  })

  it('formats KB correctly', () => {
    expect(formatFileSize(1024)).toBe('1.0 KB')
    expect(formatFileSize(1536)).toBe('1.5 KB')
  })

  it('formats MB correctly', () => {
    expect(formatFileSize(1048576)).toBe('1.0 MB')
  })

  it('formats GB correctly', () => {
    expect(formatFileSize(1073741824)).toBe('1.0 GB')
  })
})

describe('formatDate', () => {
  it('returns empty string for empty input', () => {
    expect(formatDate('')).toBe('')
  })

  it('formats a valid date string', () => {
    const result = formatDate('2024-06-15T14:30:00Z')
    // 应包含日期和时间
    expect(result).toMatch(/\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}/)
  })
})

describe('getDocTypeIcon', () => {
  it('returns folder icon for folder type', () => {
    const result = getDocTypeIcon('folder')
    expect(result.icon).toBe('Folder')
    expect(result.color).toBe('#f5a623')
  })

  it('returns doc icon for doc type', () => {
    const result = getDocTypeIcon('doc')
    expect(result.icon).toBe('Document')
    expect(result.color).toBe('#3370ff')
  })

  it('returns sheet icon for sheet type', () => {
    const result = getDocTypeIcon('sheet')
    expect(result.icon).toBe('Grid')
    expect(result.color).toBe('#36b37e')
  })

  it('returns default icon for unknown type', () => {
    const result = getDocTypeIcon('unknown')
    expect(result.icon).toBe('Document')
    expect(result.color).toBe('#999')
  })
})
