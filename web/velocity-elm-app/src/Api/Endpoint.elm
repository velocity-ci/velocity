module Api.Endpoint
    exposing
        ( CollectionOptions
        , Endpoint
        , branches
        , builds
        , commits
        , fromString
        , knownHosts
        , login
        , projectSync
        , projects
        , request
        , unwrap
        , task
        , tasks
        , toWs
        )

import Http
import Json.Decode as Decode exposing (Decoder)
import Project.Branch.Name
import Project.Commit.Hash
import Project.Slug
import Task exposing (Task)
import Url.Builder exposing (QueryParameter, int, string)
import Username exposing (Username)


{-| Http.request, except it takes an Endpoint instead of a Url.
-}
request :
    { body : Http.Body
    , expect : Http.Expect a
    , headers : List Http.Header
    , method : String
    , timeout : Maybe Float
    , url : Endpoint
    , withCredentials : Bool
    }
    -> Cmd a
request config =
    let
        req =
            if config.withCredentials then
                Http.riskyRequest
            else
                Http.request
    in
        req
            { body = config.body
            , expect = config.expect
            , headers = config.headers
            , method = config.method
            , timeout = config.timeout
            , url = unwrap config.url
            , tracker = Nothing
            }


task :
    { body : Http.Body
    , resolver : Http.Resolver e a
    , headers : List Http.Header
    , method : String
    , timeout : Maybe Float
    , url : Endpoint
    , withCredentials : Bool
    }
    -> Task e a
task config =
    let
        req =
            if config.withCredentials then
                Http.riskyTask
            else
                Http.task
    in
        req
            { body = config.body
            , resolver = config.resolver
            , headers = config.headers
            , method = config.method
            , timeout = config.timeout
            , url = unwrap config.url
            }



-- TYPES


{-| Get a URL to the Conduit API.

This is not publicly exposed, because we want to make sure the only way to get one of these URLs is from this module.

-}
type Endpoint
    = Endpoint String


unwrap : Endpoint -> String
unwrap (Endpoint str) =
    str


url : Endpoint -> List String -> List QueryParameter -> Endpoint
url baseUrl paths queryParams =
    -- NOTE: Url.Builder takes care of percent-encoding special URL characters.
    -- See https://package.elm-lang.org/packages/elm/url/latest/Url#percentEncode
    Url.Builder.crossOrigin (unwrap baseUrl)
        ("v1" :: paths)
        queryParams
        |> Endpoint


toWs : Endpoint -> String
toWs (Endpoint apiUrlBase) =
    if String.startsWith "http" apiUrlBase then
        "ws" ++ String.dropLeft 4 apiUrlBase
    else
        -- Not an http API URL - this will fail pretty quickly
        apiUrlBase



-- ENDPOINTS


login : Endpoint -> Endpoint
login baseUrl =
    url baseUrl [ "auth" ] []


type alias CollectionOptions =
    { amount : Int
    , page : Int
    }


branches : Maybe CollectionOptions -> Endpoint -> Project.Slug.Slug -> Endpoint
branches opts baseUrl projectSlug =
    url baseUrl
        [ "projects"
        , Project.Slug.toString projectSlug
        , "branches"
        ]
        (collectionParams opts)


commits : Maybe CollectionOptions -> Endpoint -> Project.Slug.Slug -> Project.Branch.Name.Name -> Endpoint
commits opts baseUrl projectSlug branchName =
    url baseUrl
        [ "projects"
        , Project.Slug.toString projectSlug
        , "branches"
        , Project.Branch.Name.toString branchName
        , "commits"
        ]
        (collectionParams opts)


builds : Maybe CollectionOptions -> Endpoint -> Project.Slug.Slug -> Endpoint
builds opts baseUrl projectSlug =
    url baseUrl
        [ "projects"
        , Project.Slug.toString projectSlug
        , "builds"
        ]
        (collectionParams opts)


tasks : Maybe CollectionOptions -> Endpoint -> Project.Slug.Slug -> Project.Commit.Hash.Hash -> Endpoint
tasks opts baseUrl projectSlug commitHash =
    url baseUrl
        [ "projects"
        , Project.Slug.toString projectSlug
        , "commits"
        , Project.Commit.Hash.toString commitHash
        , "tasks"
        ]
        (collectionParams opts)


projects : Maybe CollectionOptions -> Endpoint -> Endpoint
projects opts baseUrl =
    url baseUrl [ "projects" ] (collectionParams opts)


projectSync : Endpoint -> Project.Slug.Slug -> Endpoint
projectSync baseUrl projectSlug =
    url baseUrl [ "projects", Project.Slug.toString projectSlug, "sync" ] []


knownHosts : Maybe CollectionOptions -> Endpoint -> Endpoint
knownHosts maybeOpts baseUrl =
    url baseUrl [ "ssh", "known-hosts" ] (collectionParams maybeOpts)


collectionParams : Maybe CollectionOptions -> List QueryParameter
collectionParams maybeOpts =
    case maybeOpts of
        Just { amount, page } ->
            [ int "amount" amount
            , int "page" page
            ]

        Nothing ->
            []



-- CHANGES


fromString : String -> Endpoint
fromString =
    Endpoint
