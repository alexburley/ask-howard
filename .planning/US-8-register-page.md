# US-8 — Register Page

## User Story

**As a** new user,  
**I want** to create an account with my email and password,  
**So that** I can access the application.

## Acceptance Criteria

1. A register form is shown when the user navigates to it from the login page.
2. Form has `email` and `password` fields with a Submit button.
3. On success, the app transitions directly to the authenticated view (no separate login step).
4. Inline error messages for known failure cases:
   - 409 (email taken): **"An account with this email already exists."**
   - 422 invalid email: **"Please enter a valid email address."**
   - 422 password too short: **"Password must be at least 8 characters."**
5. Submit button is disabled while the request is in-flight.
6. A link/button to switch back to the Login page is present.

## Out of Scope

- Password confirmation field.
- Email verification flow.

---

## Implementation Plan

### `web/src/pages/RegisterPage.tsx`

Structurally identical to `LoginPage` but calls `register()` and maps `AuthError.code` to specific messages:

```ts
const errorMessages: Record<AuthErrorCode, string> = {
  EMAIL_TAKEN: 'An account with this email already exists.',
  INVALID_EMAIL: 'Please enter a valid email address.',
  PASSWORD_TOO_SHORT: 'Password must be at least 8 characters.',
  INVALID_CREDENTIALS: 'Something went wrong. Please try again.',
  NETWORK_ERROR: 'Something went wrong. Please try again.',
}
```

## Files Changed

| File | Action |
|------|--------|
| `web/src/pages/RegisterPage.tsx` | Create |
