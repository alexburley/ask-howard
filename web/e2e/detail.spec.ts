import { expect, test } from '@playwright/test'
import { fileURLToPath } from 'url'

function uniqueEmail() {
  return `test-${Date.now()}@example.com`
}

const PASSWORD = 'password123'
const TEST_ZIP = fileURLToPath(new URL('fixtures/test-content.zip', import.meta.url))

test.describe('Document Detail', () => {
  test('click image node opens panel with preview, metadata, and download; Esc closes it', async ({ page }) => {
    test.setTimeout(120_000)
    const email = uniqueEmail()

    await page.goto('/')
    await page.getByRole('button', { name: 'Create an account' }).click()
    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder(/password/i).fill(PASSWORD)
    await page.getByRole('button', { name: 'Create account' }).click()
    await expect(page.getByRole('button', { name: 'Sign out' })).toBeVisible()

    await page.locator('input[type="file"]').setInputFiles(TEST_ZIP)
    await expect(page.getByText(/document.*ready/)).toBeVisible({ timeout: 60000 })

    // Wait for canvas nodes
    const imageNode = page.locator('.doc-node img').first()
    await expect(imageNode).toBeVisible({ timeout: 10000 })

    // Click the node (parent .doc-node)
    await page.locator('.doc-node').first().click()

    // Panel opens
    const panel = page.locator('.detail-panel')
    await expect(panel).toBeVisible({ timeout: 5000 })

    // Image preview is visible
    await expect(panel.locator('img.detail-preview-image')).toBeVisible({ timeout: 5000 })

    // Metadata rows are present
    await expect(panel.getByText('Type')).toBeVisible()
    await expect(panel.getByText('Size')).toBeVisible()
    await expect(panel.getByText('Uploaded')).toBeVisible()

    // Download link is present
    await expect(panel.locator('a.detail-download')).toBeVisible()

    // Esc closes the panel
    await page.keyboard.press('Escape')
    await expect(panel).not.toBeVisible()
  })
})
