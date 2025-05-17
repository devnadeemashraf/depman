---
applyTo: "**"
---

Coding standards, domain knowledge, and preferences that AI should follow.

# Project Idea

- **Project Name**: depman
- **Project Thought**: A module that can be bundled along with the projects exe and it can check whether all the external dependencies, needed by the project, are present in the Local system/Environment where it is supposed to run. If not, it should install it and prepare the Environment for the applicaiton to run properly.
- **Project Description**: Idea is already described in the project thought. It should be a module that can be bundled along with the project exe and it can check whether all the external dependencies, needed by the project, are present in the Local system/Environment where it is supposed to run. If not, it should install it and prepare the Environment for the applicaiton to run properly. It should look for a specific file in the project directory, `app-dependencies.yml`, which will contain the list of dependencies needed by the project. It should also check for the version of the dependencies and install the required version if it is not present in the system. It should also check for the presence of `pip` in the system and install it if it is not present. It should also check for the presence of `python` in the system and install it if it is not present. It should also check for the presence of `git` in the system and install it if it is not present. It should also check for the presence of `conda` in the system and install it if it is not present. It should also check for the presence of `virtualenv` in the system and install it if it is not present. It should also check for the presence of `venv` in the system and install it if it is not present. The version also follows semantic versioning therefore it should also check whether the version of dependencies match the specified version and install only if it does not.
- **Project Type**: CLI tool & Library that can be used in other projects by importing methods and functions.
- **Project Language**: Go Lang
- **Project Language Version**: 1.24.3
- **Project Domain**: Software Development

# Response Expectations from AI

- If there are multiple things being asked in a single prompt, break them down into multiple stepped responses.
- Provide code snippets in the language specified in the project idea.
- Provide detailed explanations of the code snippets.
- Always provide well documented code snippets.
- Try to follow coding standards and best practices of the specified language and Software development in general such as DRY, KISS, YAGNI, SOLID principles.
- Explain concepts before providing code snippets for better understanding.
- Provide examples and use cases for better understanding.
- Always provide response for the version specified in the project idea.
