module Api.Endpoint
    exposing
        ( CollectionOptions
        , Endpoint
        , branches
        , fromString
        , knownHosts
        , login
        , projectSync
        , projects
        , request
        , toWs
        )

import Http
import Json.Decode as Decode exposing (Decoder)
import Phoenix.Channel exposing (Channel)
import Project.Slug
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
    -> Http.Request a
request config =
    Http.request
        { body = config.body
        , expect = config.expect
        , headers = config.headers
        , method = config.method
        , timeout = config.timeout
        , url = unwrap config.url
        , withCredentials = config.withCredentials
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
