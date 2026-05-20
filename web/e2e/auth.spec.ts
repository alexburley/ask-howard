import { expect, test } from '@playwright/test'

function uniqueEmail() {
  return `test-${Date.now()}@example.com`
}

const PASSWORD = 'password123'
const SHORT_PASSWORD = 'short'

test.describe('Register', () => {
  test('creates account and lands on workspace', async ({ page }) => {
    await page.goto('/')
    await page.getByRole('button', { name: 'Create an account' }).click()

    await page.getByPlaceholder('Email').fill(uniqueEmail())
    await page.getByPlaceholder(/password/i).fill(PASSWORD)
    await page.getByRole('button', { name: 'Create account' }).click()

    await expect(page.getByRole('button', { name: 'Sign out' })).toBeVisible()
  })

  test('shows error on duplicate email', async ({ page }) => {
    const email = uniqueEmail()

    await page.goto('/')
    await page.getByRole('button', { name: 'Create an account' }).click()
    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder(/password/i).fill(PASSWORD)
    await page.getByRole('button', { name: 'Create account' }).click()
    await expect(page.getByRole('button', { name: 'Sign out' })).toBeVisible()

    await page.getByRole('button', { name: 'Sign out' }).click()
    await page.getByRole('button', { name: 'Create an account' }).click()
    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder(/password/i).fill(PASSWORD)
    await page.getByRole('button', { name: 'Create account' }).click()

    await expect(page.getByText('An account with this email already exists.')).toBeVisible()
  })

  test('shows error on password too short', async ({ page }) => {
    await page.goto('/')
    await page.getByRole('button', { name: 'Create an account' }).click()

    await page.getByPlaceholder('Email').fill(uniqueEmail())
    await page.getByPlaceholder(/password/i).fill(SHORT_PASSWORD)
    await page.getByRole('button', { name: 'Create account' }).click()

    await expect(page.getByText('Password must be at least 8 characters.')).toBeVisible()
  })

  test('shows error on invalid email', async ({ page }) => {
    await page.goto('/')
    await page.getByRole('button', { name: 'Create an account' }).click()

    // Remove browser email validation so the invalid value reaches the server
    await page.locator('.auth-form').evaluate(f => f.setAttribute('novalidate', ''))
    await page.getByPlaceholder('Email').fill('not-an-email')
    await page.getByPlaceholder(/password/i).fill(PASSWORD)
    await page.getByRole('button', { name: 'Create account' }).click()

    await expect(page.getByText('Please enter a valid email address.')).toBeVisible()
  })
})

test.describe('Login', () => {
  test('signs in with valid credentials', async ({ page }) => {
    const email = uniqueEmail()

    // register first
    await page.goto('/')
    await page.getByRole('button', { name: 'Create an account' }).click()
    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder(/password/i).fill(PASSWORD)
    await page.getByRole('button', { name: 'Create account' }).click()
    await expect(page.getByRole('button', { name: 'Sign out' })).toBeVisible()

    // log out then log back in
    await page.getByRole('button', { name: 'Sign out' }).click()
    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder('Password').fill(PASSWORD)
    await page.getByRole('button', { name: 'Sign in' }).click()

    await expect(page.getByRole('button', { name: 'Sign out' })).toBeVisible()
  })

  test('shows error on wrong password', async ({ page }) => {
    const email = uniqueEmail()

    await page.goto('/')
    await page.getByRole('button', { name: 'Create an account' }).click()
    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder(/password/i).fill(PASSWORD)
    await page.getByRole('button', { name: 'Create account' }).click()
    await page.getByRole('button', { name: 'Sign out' }).click()

    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder('Password').fill('wrongpassword')
    await page.getByRole('button', { name: 'Sign in' }).click()

    await expect(page.getByText('Invalid email or password.')).toBeVisible()
  })

  test('shows error on unknown email', async ({ page }) => {
    await page.goto('/')
    await page.getByPlaceholder('Email').fill('nobody@example.com')
    await page.getByPlaceholder('Password').fill(PASSWORD)
    await page.getByRole('button', { name: 'Sign in' }).click()

    await expect(page.getByText('Invalid email or password.')).toBeVisible()
  })
})

test.describe('Logout', () => {
  test('returns to login page', async ({ page }) => {
    const email = uniqueEmail()

    await page.goto('/')
    await page.getByRole('button', { name: 'Create an account' }).click()
    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder(/password/i).fill(PASSWORD)
    await page.getByRole('button', { name: 'Create account' }).click()
    await expect(page.getByRole('button', { name: 'Sign out' })).toBeVisible()

    await page.getByRole('button', { name: 'Sign out' }).click()

    await expect(page.getByRole('button', { name: 'Sign in' })).toBeVisible()
  })
})
