# gopenApiToGraphQL

This project transforms OpenAPI schemas to GraphQL schemas.

#### Disclaimer

I was originally using [IBM`s openapi-to-graphql](https://github.com/IBM/openapi-to-graphql), and advise everyone to use it.
It is battle-tested (according to the GitHub Stars) and has probably way more edge cases covered.

### Story Time

I am currently developing an OpenAPI to GraphQL proxy. This project uses 3 awesome tools under the hood
1. [deepmap`s oapi-codegen](https://github.com/deepmap/oapi-codegen) to generate client stubs for the given OpenAPI schema
2. [IBM`s openapi-to-graphql](https://github.com/IBM/openapi-to-graphql) to generate a GraphQL schema from the given OpenAPI schema
3. [99desings` gqlgen](https://github.com/99designs/gqlgen) to generate the server stubs for the generated GraphQL schema

Number 1. and 3. are go projects, number 2. is a Node project. There lays the first reason why I decided to write an alternative.
A Node project means so much more stuff, e.g. package.json, package.lock, node_modules, npm has to be installed,
~~[npx installs everything everytime](https://stackoverflow.com/questions/49302438/why-does-npx-install-webpack-every-time)~~, ...

Number 2. and the more severe thing: it uses a different validator than oapi-codegen.
I am pretty sure it uses [IBM`s openapi-validator](https://github.com/IBM/openapi-validator) which is fairly strict,
whereas oapi-codegen uses [getkin`s kin-openapi](https://github.com/getkin/kin-openapi).
No I do have some public available OpenAPI services I want to use and generate GraphQL proxies for, and these server`s schemas
are sometime not good enough to get parsed by __openapi-to-graphql__. 

So I thought, if I would have to fork and modify __openapi-to-graphql__ anyway, I could also write it in Go and get all the JS
dependency and tooling out of my project.

### Limitations

There is probably a lot of open issues right now. I will edit it going along, everytime I found severe problems with it. 
I know that a lot can be done better, I am working on this for like 10 hours and is a more quick and dirty approach at the time.

Feel free to contribute and give me some PRs, if you want to.