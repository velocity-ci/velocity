module GitUrl exposing (GitUrl, decoder, sourceIcon)

import Element exposing (Element)
import Icon
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (custom, hardcoded, optional, required)
import Json.Encode as Encode



-- MODEL --


type alias GitUrl =
    { protocol : String
    , port_ : Maybe Int
    , resource : String
    , source : String
    , owner : String
    , pathName : String
    , fullName : String
    , href : String
    }



-- SERIALIZATION --


decoder : Decoder GitUrl
decoder =
    Decode.succeed GitUrl
        |> required "protocol" Decode.string
        |> required "port" (Decode.nullable Decode.int)
        |> required "resource" Decode.string
        |> required "source" Decode.string
        |> required "owner" Decode.string
        |> required "pathname" Decode.string
        |> required "full_name" Decode.string
        |> required "href" Decode.string


sourceIcon : GitUrl -> (Icon.Options -> Element msg)
sourceIcon { source } =
    case source of
        "github.com" ->
            Icon.github

        "gitlab.con" ->
            Icon.gitlab

        _ ->
            Icon.gitPullRequest
