package main

// hook data
type webhook struct {
	Action      string      `json:"action"`
	Repository  Repository  `json:"repository"`
	PullRequest PullRequest `json:"pull_request"`
	Sender      Sender      `json:"sender"`
	Refspec     string      `json:"ref"`
	RefType     string      `json:"ref_type"`
	Pusher      Pusher      `json:"pusher"`
	Forced      bool        `json:"forced"`
	Deleted     bool        `json:"deleted"`
	Created     bool        `json:"created"`
	CompareLink string      `json:"compare"`
	Commits     []Commit    `json:"commits"`
	Baserefspec string      `json:"base_ref"`
	Issue       Issue       `json:"issue"`
	Release     Release     `json:"release"`
	CheckRun    CheckRun    `json:"check_run"`
	CheckSuite  CheckSuite  `json:"check_suite"`
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

type Release struct {
	HtmlURL    string `json:"html_url"`
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Author     User   `json:"author"`
}

type CheckRun struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HtmlURL    string `json:"html_url"`
}

type CheckSuite struct {
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadBranch string `json:"head_branch"`
	App        App    `json:"app"`
}

type App struct {
	Name string `json:"name"`
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
