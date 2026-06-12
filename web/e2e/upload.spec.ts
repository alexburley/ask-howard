import { expect, test } from '@playwright/test'
import { fileURLToPath } from 'url'

function uniqueEmail() {
  return `test-${Date.now()}@example.com`
}

const PASSWORD = 'password123'
const TEST_ZIP = fileURLToPath(new URL('fixtures/test-content.zip', import.meta.url))

test.describe('Upload', () => {
  test('registers, uploads a zip, and sees processing then ready state', async ({ page }) => {
    test.setTimeout(90_000)
    const email = uniqueEmail()

    await page.goto('/')
    await page.getByRole('button', { name: 'Create an account' }).click()
    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder(/password/i).fill(PASSWORD)
    await page.getByRole('button', { name: 'Create account' }).click()
    await expect(page.getByRole('button', { name: 'Sign out' })).toBeVisible()

    await expect(page.getByText('Drop a zip file here or click to browse')).toBeVisible()
    await page.locator('input[type="file"]').setInputFiles(TEST_ZIP)

    // Should show uploading or processing
    await expect(page.getByText(/Uploading|Processing your documents/)).toBeVisible({ timeout: 10000 })

    // Should eventually reach ready state with document count
    await expect(page.getByText(/document.*ready/)).toBeVisible({ timeout: 60000 })
  })
})
