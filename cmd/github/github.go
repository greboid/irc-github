package main

//ping events
type pinghook struct {
}

//Push events
type pushhook struct {
	Refspec     string     `json:"ref"`
	Repository  Repository `json:"repository"`
	Pusher      Pusher     `json:"pusher"`
	Sender      Sender     `json:"sender"`
	Forced      bool       `json:"forced"`
	Deleted     bool       `json:"deleted"`
	Created     bool       `json:"created"`
	CompareLink string     `json:"compare"`
	Commits     []Commit   `json:"commits"`
	Baserefspec string     `json:"base_ref"`
}

type Repository struct {
	FullName  string `json:"full_name"`
	IsPrivate bool   `json:"private"`
}

type Pusher struct {
	Name string `json:"name"`
}

type Commit struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	URL       string `json:"url"`
	Author    Author `json:"author"`
	Committer Author `json:"committer"`
}

type Author struct {
	User string `json:"username"`
}

type Sender struct {
	Login string `json:"login"`
}

//Pull Request events
type prhook struct {
	Action      string      `json:"action"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
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
}

type User struct {
	Login string `json:"login"`
}

//Issue events
type issuehook struct {
	Action     string     `json:"action"`
	Issue      Issue      `json:"issue"`
	Repository Repository `json:"repository"`
	User       User       `json:"sender"`
}

type Issue struct {
	HtmlURL string `json:"html_url"`
	Title   string `json:"title"`
	State   string `json:"state"`
	User    User   `json:"user"`
}
