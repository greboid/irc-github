package main

//hook data
type webhook struct {
	Action      string      `json:"action"`
	Repository Repository `json:"repository"`
	PullRequest PullRequest `json:"pull_request"`
	Sender     Sender     `json:"sender"`
	Refspec     string     `json:"ref"`
	Pusher      Pusher     `json:"pusher"`
	Forced      bool       `json:"forced"`
	Deleted     bool       `json:"deleted"`
	Created     bool       `json:"created"`
	CompareLink string     `json:"compare"`
	Commits     []Commit   `json:"commits"`
	Baserefspec string     `json:"base_ref"`
	Issue      Issue      `json:"issue"`
}

type Repository struct {
	FullName  string `json:"full_name"`
	IsPrivate bool   `json:"private"`
}

type Commit struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	URL       string `json:"url"`
	Author    Author `json:"author"`
	Committer Author `json:"committer"`
}

type PullRequest struct {
	HtmlURL  string `json:"html_url"`
	Url      string `json:"url"`
	State    string `json:"state"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	User     User   `json:"user"`
	Merged   string `json:"merged_at"`
	MergedBy User   `json:"merged_by"`
	Sender   Sender `json:"sender"`
}

type Issue struct {
	HtmlURL string `json:"html_url"`
	Title   string `json:"title"`
	State   string `json:"state"`
	User    User   `json:"user"`
}

type Pusher struct {
	Name string `json:"name"`
}

type Author struct {
	User string `json:"username"`
}

type Sender struct {
	Login string `json:"login"`
}

type User struct {
	Login string `json:"login"`
}
