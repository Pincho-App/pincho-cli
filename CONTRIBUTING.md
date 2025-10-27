# Contributing to WirePusher JavaScript SDK

Thank you for your interest in contributing to the WirePusher JavaScript SDK! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When creating a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples** (code snippets, configuration, etc.)
- **Describe the behavior you observed** and what you expected
- **Include version information** (Node.js version, SDK version, OS)
- **Include error messages and stack traces** if applicable

**Example bug report:**

```markdown
## Bug: Client throws TypeError on successful response

**Environment:**
- WirePusher SDK: 1.0.0
- Node.js: 18.17.0
- OS: macOS 13.4

**Steps to Reproduce:**
1. Initialize client with valid credentials
2. Call `client.send('Title', 'Message')`
3. Observe error despite successful API response

**Expected Behavior:**
Method should resolve successfully

**Actual Behavior:**
Throws TypeError: Cannot read property 'status' of undefined

**Stack Trace:**
```
TypeError: Cannot read property 'status' of undefined
    at WirePusher.send (client.ts:129)
    ...
```

### Suggesting Enhancements

Enhancement suggestions are tracked as GitLab issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide a detailed description** of the proposed enhancement
- **Explain why this enhancement would be useful** to most users
- **Provide examples** of how the enhancement would be used
- **List any alternatives** you've considered

### Pull Requests

1. Fork the repository
2. Create a new branch from `main`:
   ```bash
   git checkout -b feature/my-new-feature
   ```
3. Make your changes following the coding standards
4. Add or update tests as needed
5. Ensure all tests pass
6. Update documentation if needed
7. Commit your changes with clear commit messages
8. Push to your fork
9. Open a Merge Request against `main`

## Development Setup

### Prerequisites

- Node.js >= 18.0.0
- npm >= 8.0.0
- Git

### Setting Up Your Development Environment

1. **Clone the repository:**
   ```bash
   git clone https://gitlab.com/wirepusher/javascript-sdk.git
   cd javascript-sdk
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Verify the setup:**
   ```bash
   npm test
   npm run build
   ```

### Project Structure

```
javascript-sdk/
├── src/              # Source code
│   ├── client.ts     # Main WirePusher client
│   ├── errors.ts     # Custom error classes
│   ├── types.ts      # TypeScript type definitions
│   └── index.ts      # Public API exports
├── tests/            # Test files
│   ├── client.test.ts
│   └── errors.test.ts
├── examples/         # Usage examples
├── dist/             # Built package (generated)
└── coverage/         # Test coverage reports (generated)
```

## Coding Standards

### TypeScript

- **Use TypeScript strict mode** - All code must pass strict type checking
- **Provide proper type annotations** for public APIs
- **Use meaningful variable names** - Prefer descriptive names over short abbreviations
- **Keep functions focused** - Each function should do one thing well

### Code Style

We use ESLint and Prettier to enforce consistent code style:

```bash
# Check linting
npm run lint

# Auto-fix linting issues
npm run lint:fix

# Check formatting
npm run format:check

# Auto-format code
npm run format
```

**Key style points:**
- Use 2 spaces for indentation
- Use single quotes for strings
- No semicolons (except where required)
- Trailing commas in multi-line structures
- Max line length: 100 characters (soft limit)

### Error Handling

- Always use the custom error classes (`WirePusherError`, `WirePusherAuthError`, `WirePusherValidationError`)
- Provide clear error messages that help users understand what went wrong
- Include context in error messages when helpful
- Use proper error chaining when wrapping errors

**Example:**

```typescript
try {
  const response = await fetch(url);
  if (!response.ok) {
    throw new WirePusherAuthError('Invalid credentials');
  }
} catch (error) {
  if (error instanceof WirePusherError) {
    throw error;
  }
  throw new WirePusherError(`Network error: ${error.message}`);
}
```

### Documentation

- **Add JSDoc comments** for all public APIs
- **Include examples** in JSDoc comments
- **Keep README.md up to date** with API changes
- **Update CHANGELOG.md** for all user-facing changes

**Example:**

```typescript
/**
 * Send a notification via WirePusher API.
 *
 * @param title - Notification title
 * @param message - Notification message
 * @returns Promise resolving to the API response
 * @throws {WirePusherAuthError} Invalid credentials
 * @throws {WirePusherValidationError} Invalid parameters
 *
 * @example
 * ```typescript
 * await client.send('Build Complete', 'v1.2.3 deployed');
 * ```
 */
async send(title: string, message: string): Promise<NotificationResponse>
```

## Testing

### Writing Tests

- All new features must include tests
- Aim for >90% code coverage
- Use descriptive test names
- Test both success and error cases
- Test edge cases and boundary conditions

**Test structure:**

```typescript
describe('Feature', () => {
  describe('method', () => {
    it('should handle success case', async () => {
      // Arrange
      // Act
      // Assert
    });

    it('should handle error case', async () => {
      // Arrange
      // Act
      // Assert
    });
  });
});
```

### Running Tests

```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage
```

### Coverage Requirements

- **Statements:** >= 90%
- **Branches:** >= 90%
- **Functions:** >= 90%
- **Lines:** >= 90%

## Building

```bash
# Build the package
npm run build

# Type check without building
npm run typecheck
```

The build process creates:
- `dist/index.js` - ES modules build
- `dist/index.cjs` - CommonJS build
- `dist/index.d.ts` - TypeScript definitions (ESM)
- `dist/index.d.cts` - TypeScript definitions (CJS)

## Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Build process or auxiliary tool changes

**Examples:**

```
feat(client): add retry logic for failed requests

fix(errors): correct error message for validation failures

docs(readme): update installation instructions

test(client): add tests for timeout handling
```

## Merge Request Process

1. **Update documentation** if you've changed APIs
2. **Update CHANGELOG.md** with your changes
3. **Ensure all tests pass** and coverage meets requirements
4. **Ensure linting passes** with no warnings
5. **Request review** from maintainers
6. **Address review feedback** promptly
7. **Squash commits** if requested before merging

### Merge Request Checklist

- [ ] Tests added/updated and passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Code follows style guidelines
- [ ] No linting warnings
- [ ] TypeScript compiles without errors
- [ ] All checks passing in CI/CD

## Release Process

(For maintainers)

1. Update version in `package.json`
2. Update `CHANGELOG.md` with release notes
3. Commit changes: `git commit -am "chore: release v1.x.x"`
4. Create Git tag: `git tag v1.x.x`
5. Push changes and tag: `git push origin main --tags`
6. Cloud Build will automatically publish to npm

## Getting Help

- **GitLab Issues:** https://gitlab.com/wirepusher/javascript-sdk/-/issues
- **Email:** support@wirepusher.com
- **Documentation:** https://wirepusher.com/api

## Recognition

Contributors will be recognized in:
- The project's README.md
- Release notes in CHANGELOG.md
- GitLab's contribution graphs

Thank you for contributing to WirePusher!
