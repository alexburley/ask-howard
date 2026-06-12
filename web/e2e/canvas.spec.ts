import { expect, test } from '@playwright/test'
import { fileURLToPath } from 'url'

function uniqueEmail() {
  return `test-${Date.now()}@example.com`
}

const PASSWORD = 'password123'
const TEST_ZIP = fileURLToPath(new URL('fixtures/test-content.zip', import.meta.url))

test.describe('Canvas', () => {
  test('shows documents as nodes on the canvas after upload', async ({ page }) => {
    test.setTimeout(120_000)
    const email = uniqueEmail()

    await page.goto('/')
    await page.getByRole('button', { name: 'Create an account' }).click()
    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder(/password/i).fill(PASSWORD)
    await page.getByRole('button', { name: 'Create account' }).click()
    await expect(page.getByRole('button', { name: 'Sign out' })).toBeVisible()

    await expect(page.getByText('Drop a zip file here or click to browse')).toBeVisible()
    await page.locator('input[type="file"]').setInputFiles(TEST_ZIP)
    await expect(page.getByText(/document.*ready/)).toBeVisible({ timeout: 60000 })

    // Canvas should appear with at least one document node
    await expect(page.locator('.doc-node').first()).toBeVisible({ timeout: 10000 })

    // At least one image node should render an <img>
    await expect(page.locator('.doc-node img').first()).toBeVisible()
  })
})
