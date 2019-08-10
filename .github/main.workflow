workflow "Build go" {
  resolves = ["build"]
  on = "push"
}

action "build" {
  uses = "cedrickring/golang-action/go1.12@master"
}

