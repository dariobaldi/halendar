package main

import "testing"

// TestRoutesDoNotPanic guards against a real production incident: httprouter builds
// a separate radix tree per HTTP method, but panics at startup -- not at build or vet
// time -- if any one method's tree has both a static and a wildcard segment at the
// same position (e.g. PUT "/v1/ai-settings/provider" alongside PUT
// "/v1/ai-settings/:provider/key"). Registering routes.go's table once here catches
// that class of bug before it reaches a deploy.
func TestRoutesDoNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("app.routes() panicked (this is httprouter rejecting a route conflict -- see the panic message for which path): %v", r)
		}
	}()

	(&app{}).routes()
}
