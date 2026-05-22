import { expect, test } from '@playwright/test'
import { join } from 'path'

function uniqueEmail() {
  return `test-${Date.now()}@example.com`
}

const PASSWORD = 'password123'
const TEST_ZIP = join(__dirname, 'fixtures', 'test.zip')

test.describe('Upload', () => {
  test('registers, uploads a zip, and sees processing state', async ({ page }) => {
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

    await expect(page.getByText(/Uploading|Processing|Upload complete/)).toBeVisible({ timeout: 15000 })
  })
})
