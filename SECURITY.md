# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability in Devgraph CLI, please report it privately.

### How to Report

1. **Email**: Send details to security@arctir.com
2. **Include**:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Expect

- **Acknowledgment**: Within 24 hours
- **Initial Assessment**: Within 72 hours
- **Resolution**: Security fixes are prioritized and typically released within 7 days for critical issues

### Security Response Process

1. **Triage**: We assess the severity and impact
2. **Fix Development**: We develop and test a fix
3. **Disclosure**: We coordinate disclosure with the reporter
4. **Release**: Security fix is released with advisory

## Security Measures

### Authentication & Authorization
- Uses OIDC (OpenID Connect) for secure authentication
- JWT tokens with proper validation
- Automatic token refresh with secure storage
- Environment-based access controls

### Data Protection
- Credentials stored in OS-specific secure locations
- No sensitive data in logs
- HTTPS for all API communications
- Private module handling with secure Git configuration

### Code Security
- **Static Analysis**: Gosec, CodeQL, and GoLint security rules
- **Dependency Scanning**: Govulncheck for known vulnerabilities
- **Secret Scanning**: TruffleHog for committed secrets
- **License Compliance**: Automated license checking
- **SARIF Reports**: Security findings uploaded to GitHub Security tab

### CI/CD Security
- **Pinned Actions**: All GitHub Actions use specific versions
- **Minimal Permissions**: Workflows use least-privilege principle
- **Secret Management**: Secure handling of build secrets
- **Supply Chain**: Dependency review on all PRs

### Development Security
- **Security Linting**: Comprehensive golangci-lint configuration
- **Error Handling**: Mandatory error checking with errcheck
- **Input Validation**: All user inputs validated
- **Safe Defaults**: Secure-by-default configuration

## Security Testing

Run security scans locally:

```bash
# Run all security checks
make security

# Individual security scans
make vuln-check          # Vulnerability scanning
golangci-lint run        # Static analysis with security rules
go install github.com/securecodewarrior/gosec/cmd/gosec@latest
gosec ./...              # Gosec security scanner
```

## Security Best Practices

### For Users
- Keep CLI updated to latest version
- Verify binary checksums when downloading
- Use environment-specific configurations
- Regularly rotate authentication tokens
- Report suspicious behavior immediately

### For Developers
- Follow secure coding practices
- Run security scans before committing
- Never commit secrets or credentials
- Use parameterized queries for any database operations
- Validate all inputs and sanitize outputs
- Implement proper error handling

## Known Security Considerations

### OAuth2/OIDC Flow
- Uses local redirect server on `localhost:8080`
- Tokens stored in OS keychain/credential manager
- PKCE (Proof Key for Code Exchange) used for additional security

### Network Communications
- All API calls use HTTPS
- Certificate validation enforced
- Timeout configurations prevent hanging requests

### File System Access
- Configuration files have restricted permissions
- Temporary files cleaned up after use
- Path traversal protection implemented

## Security Updates

Security updates are released as patch versions and communicated through:
- GitHub Security Advisories
- Release notes with `[SECURITY]` prefix
- Email notifications to registered users (if applicable)

## Contact

For security-related questions or concerns:
- **Security Team**: security@arctir.com
- **General Support**: support@arctir.com

## Acknowledgments

We appreciate security researchers who responsibly disclose vulnerabilities and help improve Devgraph CLI's security posture.
