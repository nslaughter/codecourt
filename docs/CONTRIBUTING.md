# Contributing to CodeCourt

Thank you for your interest in contributing to CodeCourt! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Please read it before contributing.

## Getting Started

### Issues

- Check existing issues to see if your problem or idea has already been addressed.
- For bugs, please provide detailed steps to reproduce, error messages, and your environment details.
- For features, please describe the use case and expected behavior.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature-name`)
3. Make your changes
4. Run tests and linting (`make test && make lint`)
5. Commit your changes (see Commit Guidelines below)
6. Push to your branch (`git push origin feature/your-feature-name`)
7. Open a Pull Request

## Development Environment

Please refer to the [DEVELOPMENT.md](DEVELOPMENT.md) guide for detailed instructions on setting up your development environment.

## Coding Standards

### Go Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code (`make fmt`)
- Pass `golangci-lint` checks (`make lint`)
- Write idiomatic Go code

### Testing

- Write table-driven tests for all new code
- Maintain or improve test coverage
- Include unit tests for all new functionality
- Add integration tests for service interactions
- Include end-to-end tests for new endpoints or features

### Documentation

- Update documentation for any changed functionality
- Document all exported functions, types, and constants
- Add examples for complex functionality
- Keep documentation up-to-date with code changes

## Commit Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: Code changes that neither fix a bug nor add a feature
- `perf`: Performance improvements
- `test`: Adding or fixing tests
- `chore`: Changes to the build process or auxiliary tools

### Scope

The scope should be the name of the package or service affected:

- `api-gateway`
- `user-service`
- `problem-service`
- `submission-service`
- `judging-service`
- `notification-service`
- `k8s` (for Kubernetes/deployment changes)
- `docs` (for documentation)

### Subject

- Use the imperative, present tense: "change" not "changed" nor "changes"
- Don't capitalize the first letter
- No period (.) at the end

### Examples

```
feat(problem-service): add support for multiple test cases

fix(judging-service): resolve memory limit issue in container execution

docs(readme): update deployment instructions
```

## Pull Request Process

1. Ensure your code passes all tests and lint checks
2. Update documentation as needed
3. Include tests for new functionality
4. Link any relevant issues
5. Get approval from at least one maintainer
6. Squash commits if requested

## Code Review

All submissions require review. We use GitHub pull requests for this purpose.

During code review, maintainers will look for:

- Adherence to the coding standards
- Test coverage
- Documentation
- Performance implications
- Security considerations

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run end-to-end tests
make e2e-test
```

### Writing Tests

- Use table-driven tests for Go code
- Mock external dependencies
- Test edge cases and error conditions
- Keep tests fast and deterministic

## Kubernetes and Helm

For changes to Kubernetes resources or Helm charts:

1. Test changes locally with Kind
2. Verify that the chart installs cleanly
3. Test upgrade scenarios
4. Document any new values or configuration options

## Release Process

The maintainers will handle the release process, which includes:

1. Version bumping
2. Changelog generation
3. Docker image building and pushing
4. Helm chart packaging and publishing

## Getting Help

If you need help with your contribution:

- Ask questions in the issue you're working on
- Reach out to maintainers
- Check the documentation

Thank you for contributing to CodeCourt!
