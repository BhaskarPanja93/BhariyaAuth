package sso

import (
	CookieProcessor "BhariyaAuth/processors/cookies"
	ResponseProcessor "BhariyaAuth/processors/sso"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"

	"github.com/gofiber/fiber/v3"
	"github.com/markbates/goth"
)

// Step1 initializes the Single Sign-On (SSO) authentication flow.
//
// This function acts as the entry point for initiating authentication with an external SSO provider.
// It performs the following responsibilities:
//  1. Extracts the requested SSO provider from the route parameters.
//  2. Validates and retrieves the provider configuration from Goth.
//  3. Constructs a secure "state" payload containing metadata required across the SSO flow.
//  4. Serializes and encrypts this state to prevent tampering.
//  5. Initiates a session with the provider using the encrypted state.
//  6. Stores the session securely in a cookie (encrypted).
//  7. Redirects the user to the provider's authentication URL.
//
// Security Considerations:
// - The state object is encrypted before being sent to the provider to prevent manipulation.
// - Session data is also encrypted before being stored in cookies.
// - Expiry is embedded into the state to enforce session validity.
//
// Query Parameters:
// - `remember=yes|no`: Determines whether the session should persist beyond default duration.
//
// Route Parameters:
// - `/sso/:provider`: Specifies the SSO provider (must match Goth provider names).
//
// Returns:
// - Redirect response to the SSO provider on success.
// - Error response popup on failure at any step.
func Step1(ctx fiber.Ctx) error {

	// Extract provider name from URL parameters (must match registered Goth providers)
	providerName := ctx.Params(ProviderParam)

	// Attempt to retrieve the SSO provider configuration from Goth
	provider, err := goth.GetProvider(providerName)
	if err != nil {
		// Fail early if provider is not recognized
		return ResponseProcessor.FailurePopup(ctx, UnknownProvider)
	}

	// Construct the SSO state payload that will persist across the authentication flow
	// This state is later returned by the provider and used for validation
	encryptedState, err := TokenProcessor.CreateSSOToken(ctx, providerName)
	if err != nil {
		// Abort if encryption fails (critical security failure)
		return ResponseProcessor.FailurePopup(ctx, StateEncryptFailed)
	}

	// Initialize the authentication session with the provider using encrypted state
	// This state is typically sent as a query parameter to the provider
	session, err := provider.BeginAuth(encryptedState)
	if err != nil {
		// Abort if provider session initialization fails
		return ResponseProcessor.FailurePopup(ctx, SessionCreateFailed)
	}

	// Retrieve the authentication URL to which the user must be redirected
	authURL, err := session.GetAuthURL()
	if err != nil {
		// Abort if URL generation fails
		return ResponseProcessor.FailurePopup(ctx, AuthURLNotFound)
	}

	// Encrypt the session data before storing it in the client cookie
	encryptedSession, err := StringProcessor.EncryptInterfaceToString(session)
	if err != nil {
		// Abort if session serialization fails
		return ResponseProcessor.FailurePopup(ctx, SessionEncryptFailed)
	}

	// Attach the encrypted session as an SSO cookie in the response
	// This allows retrieval during the callback phase of the SSO flow
	CookieProcessor.AttachSSOCookie(ctx, encryptedSession)

	// Redirect the user to the provider's authentication page
	// The encrypted state is passed along and session is maintained via cookie
	return ctx.Redirect().To(authURL)
}
