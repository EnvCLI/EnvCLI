# CI Integration

EnvCLI automatically detects execution in CI environments based on the env variable (CI=true) and will pass all variables into each container you use - so you can use variables like GITLAB_ or a BINTRAY_AUTH_TOKEN within the containers.
