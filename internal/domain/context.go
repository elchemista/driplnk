package domain

type contextKey string

const (
	CtxKeyTargetUserID contextKey = "target_user_id"
	CtxKeyUser         contextKey = "user"    // Logged in user
	CtxKeySession      contextKey = "session" // Session data
)
