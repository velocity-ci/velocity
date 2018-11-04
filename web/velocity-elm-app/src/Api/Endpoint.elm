module Api.Endpoint exposing (Endpoint, fromString, login, projects, request)

import Http
import Json.Decode as Decode exposing (Decoder)
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



-- ENDPOINTS


login : Endpoint -> Endpoint
login baseUrl =
    url baseUrl [ "auth" ] []


projects : Int -> Endpoint -> Endpoint
projects amount baseUrl =
    url baseUrl [ "projects" ] [ int "amount" amount ]



-- CHANGES


fromString : String -> Endpoint
fromString =
    Endpoint
