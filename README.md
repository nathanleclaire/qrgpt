# qrgpt

![image](https://user-images.githubusercontent.com/1476820/209900713-235b65c0-561f-4b23-b434-69a4d1e76572.png)

[ChatGPT](https://chat.openai.com/) is an excellent tool, but it functions best when
given specific instruction such as app-specific concepts, including database table schemas,
existing code and libraries, etc. Otherwise, even if it spits out code that is generally 
correct, transposing it to your own use cases will be a pain. `qrgpt` ("Query GPT") is a 
CLI utility to help with this.

## Usage

Currently, qrgpt is just a simple Go binary:

```
$ go install github.com/nathanleclaire/qrgpt
```

Define a config file, `~/.qrgpt` with a prompt containing a template for usage with a given
repo. Currently, only one prompt is supported, but I'd like to support more.

For instance, maybe you want some information about your app, including database schemas,
and instructions to GPT to not include structs, etc., that it will tend to spit out otherwise.
Using the `exec` template function, you can even execute commands. 
All [sprig](http://masterminds.github.io/sprig/) template functions are included as well.

Hence, you can write something like this.

```yaml
version: "1"
repos:
  github.com/yourgithubuser/app:
    prompt: >
      I have a sqlite3 schema like this

      {{ range $i, $table := index .Args 1 | split "," }}

      {{ $schema := printf ".schema %s" $table }}
      {{ exec "sqlite3" "./backend/db.sqlite3" $schema }}

      {{end -}}

      - Write me a query to {{index .Args 2}} with jmoiron/sqlx.

      - Do not include the sqlx.Open, I already have an instance, db.

      - Do not include struct definitions for the tables.

      - Use logrus.WithFields("error", err).Error to log it and then do an http.Error.

      - Omit comments.
```

Then, from within the repo locally, if `main` or `master` matches the defined repo remote (YAML key),
you can run `qrgpt` in that directory with arguments to get a fully contexted Chat-GPT query:

```
$ qrgpt 
$ qrgpt accounts,samples "Get the accounts with the most samples"                                                           
I have a sqlite3 schema like this

CREATE TABLE accounts (
  -- Full definition omitted for brevity
);


CREATE TABLE samples (
  -- Full definition omitted for brevity
);

- Write me a query to Get the accounts with the most samples with jmoiron/sqlx.
- Do not include the sqlx.Open, I already have an instance, db.
- Do not include struct definitions for the tables.
- Use logrus.WithFields("error", err).Error to log it and then do an http.Error.
- Omit comments.
```

Which can easily be copypasted (or piped to `pbcopy` to plug into ChatGPT.

<img width="599" alt="image" src="https://user-images.githubusercontent.com/1476820/209900563-fd751c17-122b-4a08-8ec8-c73267548360.png">

Notice that ChatGPT has produced more usable output because it has adhered to our specific libraries
and preferences -- it did not regurgitate redundant structs or comments, and it handled errors as 
requested.
