workflow "Build go" {
  resolves = ["build"]
  on = "push"
}

action "build" {
  uses = "sosedoff/actions/golang-build@master"

  // Optional args for specific architechtures
  args = "linux/amd64"
}

