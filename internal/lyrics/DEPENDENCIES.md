Chinese catalogue matching uses github.com/yanmingcao/opencc-go v1.0.0
(Apache-2.0, declared in source headers and README), pinned in go.mod/go.sum.
Go and x/text Unicode normalization do not convert Chinese scripts. Reuse
OpenCC's existing official dictionaries rather than a partial character table.
Embedded dictionaries require no installation or C runtime; minimum Go 1.21
is compatible with the project's Go 1.26. Conversion runs only during lookup.
The library has no runtime third-party dependencies.
The alternative longbridgeapp/opencc v0.3.13 was rejected because its cedar-go
transitive dependency ships a GPL-2.0 license, unsuitable for this MIT project.
