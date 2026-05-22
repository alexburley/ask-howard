import { expect, test } from '@playwright/test'
import { join } from 'path'

function uniqueEmail() {
  return `test-${Date.now()}@example.com`
}

const PASSWORD = 'password123'
const TEST_ZIP = join(__dirname, 'fixtures', 'test-content.zip')

test.describe('Upload', () => {
  test('registers, uploads a zip, and sees processing then ready state', async ({ page }) => {
    const email = uniqueEmail()

    await page.goto('/')
    await page.getByRole('button', { name: 'Create an account' }).click()
    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder(/password/i).fill(PASSWORD)
    await page.getByRole('button', { name: 'Create account' }).click()
    await expect(page.getByRole('button', { name: 'Sign out' })).toBeVisible()

    const [fileChooser] = await Promise.all([
      page.waitForEvent('filechooser'),
      page.getByText('Drop a zip file here or click to browse').click(),
    ])
    await fileChooser.setFiles(TEST_ZIP)

    // Should show uploading or processing
    await expect(page.getByText(/Uploading|Processing your documents/)).toBeVisible({ timeout: 10000 })

    // Should eventually reach ready state with document count
    await expect(page.getByText(/document.*ready/)).toBeVisible({ timeout: 30000 })
  })
})
