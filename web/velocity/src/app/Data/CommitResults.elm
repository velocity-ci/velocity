module Data.CommitResults exposing (..)

import Data.Commit as Commit exposing (Commit)
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (decode, required)


type alias Results =
    { results : List Commit
    , total : Int
    }



-- SERIALIZATION --


decoder : Decoder Results
decoder =
    decode Results
        |> required "result" (Decode.list Commit.decoder)
        |> required "total" Decode.int
