module Data.Project exposing (Project, decoder)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (custom, decode, hardcoded, required)


type alias Project =
    { id : String
    , key : String
    , name : String
    , repository : String
    , createdAt : String
    , updatedAt : String
    }



-- SERIALIZATION --


decoder : Decoder Project
decoder =
    decode Project
        |> required "id" Decode.string
        |> required "key" Decode.string
        |> required "name" Decode.string
        |> required "repository" Decode.string
        |> required "createdAt" Decode.string
        |> required "updatedAt" Decode.string
