module Data.KnownHost exposing (KnownHost, decoder)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (custom, decode, hardcoded, required)

type alias KnownHost =
    { hosts: List String
    , comment: String
    , sha256: String
    , md5: String
    }


-- SERIALIZATION --

decoder : Decoder KnownHost
decoder =
    decode KnownHost
        |> required "hosts" (Decode.list Decode.string)
        |> required "comment" Decode.string
        |> required "sha256" Decode.string
        |> required "md5" Decode.string
