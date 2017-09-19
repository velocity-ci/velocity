module Data.Commit exposing (..)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (decode, required)
import Time.DateTime as DateTime exposing (DateTime)
import Data.Helpers exposing (stringToDateTime)


type alias Commit =
    { hash : String
    , author : String
    , date : DateTime
    , message : String
    }



-- SERIALIZATION --


decoder : Decoder Commit
decoder =
    decode Commit
        |> required "hash" Decode.string
        |> required "author" Decode.string
        |> required "date" stringToDateTime
        |> required "message" Decode.string
