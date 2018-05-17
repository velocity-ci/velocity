module Data.GitUrl exposing (GitUrl, decoder)

-- EXTERNAL --

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (custom, decode, hardcoded, required, optional)


-- MODEL --


type alias GitUrl =
    { protocols : List String
    , port_ : Maybe Int
    , resource : String
    , source : String
    }



-- SERIALIZATION --


decoder : Decoder GitUrl
decoder =
    decode GitUrl
        |> required "protocols" (Decode.list Decode.string)
        |> required "port" (Decode.nullable Decode.int)
        |> required "resource" Decode.string
        |> required "source" Decode.string
