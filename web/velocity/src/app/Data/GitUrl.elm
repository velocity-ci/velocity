module Data.GitUrl exposing (GitUrl, decoder)

-- EXTERNAL --

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (custom, decode, hardcoded, optional, required)


-- MODEL --


type alias GitUrl =
    { protocol : String
    , port_ : Maybe Int
    , resource : String
    , source : String
    }



-- SERIALIZATION --


decoder : Decoder GitUrl
decoder =
    decode GitUrl
        |> required "protocol" Decode.string
        |> required "port" (Decode.nullable Decode.int)
        |> required "resource" Decode.string
        |> required "source" Decode.string
