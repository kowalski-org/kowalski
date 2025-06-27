# Kowalski
Kowalski is an AI which helps to configure your system. It's basic idea is
store the openSUSE documentation in a vector datbase and if the user has
a request to `kowalksi`, the relevant documents are retreived from the database
and presented to an LLM which is running on ollama as context.
If files are mentioned in the documentation these files are also presented to the LLM.
In opposited to MCP this is a single shot approach, as the LLM know so immediatelty of how to start.

# Installation

## RPM

RPM packages are available in [OBS](https://build.opensuse.org/package/show/science:machinelearning/kowalski) in the `science:machinelearning` project.

[!NOTE]

Make sure that an ollama instance is started in a seperate terminal with

```
  ollama serve
```

## Container

You can run kowalksi as container which is available at the [gthub registry](https://github.com/users/mslacken/packages/container/package/kowalski-binary)

## Source

You need following packages installed
* `go`
* `ollama`
* `faiss-devel` from `science:machinelearning`

After this you will have to dowload the go dependencies with
```
  go mod vendor
```
Now you have to start the ollama service with
```
  ollama serve
```
and in another terminal pull the needed models with
```
  ollama pull gemma3:4b
  ollama pull nomic-embed-text:v1.5
```
After this you need to create a knowledge database, for this clone the suse docs with
```
  git clone https://github.com/SUSE/doc-modular.git
```
and initialize it with
```
  find PATHTOSUSEDOCS -name \*xml -type f  | xargs go run main.go --database ./kwDB database add susedoc@nomic-embed-text:v1.5
```
Finally you can open the chat with
```
  go run main.go chat
```
what should give something like: ![Screenshot of chat](./Startsshd.png)

