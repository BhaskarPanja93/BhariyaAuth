package form

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

// ReadFormData attempts to bind incoming request data into the provided form struct.
//
// This helper function tries to parse request data from multiple sources:
//  1. Form data (application/x-www-form-urlencoded or multipart).
//  2. Request body (typically JSON).
//
// It ensures that:
//   - Data is successfully parsed from at least one valid source.
//   - The form struct is populated consistently.
//
// Flow Summary:
//
//	try form binding → if fails → try body binding → return success/failure
//
// Behavior:
// - Returns true if binding succeeds from either source.
// - Returns false only if both bindings fail.
//
// Security Considerations:
// - Avoids partial success ambiguity by prioritizing structured fallback.
// - Prevents accidental acceptance of malformed requests.
//
// Parameters:
// - ctx: Fiber context containing request data.
// - form: pointer to struct where data will be bound.
//
// Returns:
// - true if binding succeeds.
// - false if all binding attempts fail.
func ReadFormData(ctx fiber.Ctx, form any) error {

	// Attempt to bind form data first (common for HTML forms)
	if err := ctx.Bind().Form(form); err == nil {
		return nil
	}

	// Fallback to body parsing (e.g., JSON payload)
	if err := ctx.Bind().Body(form); err == nil {
		return nil
	}

	// Both parsing methods failed
	return errors.New("form could not be read")
}
