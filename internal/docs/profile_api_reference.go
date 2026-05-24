package docs

import "fmt"

type apiReferenceProfile struct{}

func (apiReferenceProfile) Name() string { return "api-reference" }
func (apiReferenceProfile) Description() string {
	return "API reference style with endpoints, parameters, and response formats"
}

func (apiReferenceProfile) BuildPrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are an API documentation specialist. Create API reference documentation for the following source code.

For each exported type and function:
- Signature (name, parameters, return values)
- Description of what it does
- Parameter descriptions
- Return value descriptions
- Example usage
- Error conditions

Use structured formatting suitable for API docs.

File: %s

%s`, file, string(content))
}

func (apiReferenceProfile) BuildUpdatePrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are an API documentation specialist. Update the API reference for the following changed file.

Focus on:
- Updated signatures (new params, changed returns)
- New exported types or functions
- Removed or deprecated items
- Updated examples

File: %s

%s`, file, string(content))
}

func (apiReferenceProfile) FormatOutput(doc string) string { return doc }

func init() {
	RegisterProfile(apiReferenceProfile{})
}
