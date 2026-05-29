package client

type Task struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	StarterCode  string `json:"starter_code"`
	Code         string `json:"code"`
	Language     string `json:"language"`
	Status       string `json:"status"`
	Grade        string `json:"grade"`
	Feedback     string `json:"feedback"`
	Archived     bool   `json:"archived"`
	User         string `json:"user"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StarterCode string `json:"starter_code"`
	Language    string `json:"language"`
}

type UpdateTaskRequest struct {
	Code string `json:"code"`
}

type GradeRequest struct {
	Grade    string `json:"grade"`
	Feedback string `json:"feedback"`
}

type AuthRequest struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token  string `json:"token"`
	Record any    `json:"record"`
}

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Created   string `json:"created"`
	Updated   string `json:"updated"`
}

type ListTasksResponse []Task