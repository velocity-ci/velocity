module Data.LogOutput exposing (..)


type alias LogOutput =
    { step : Int
    , status : Status
    , commit : Commit.Hash
    , project : Project.Id
    , task : Task.Name
    }



-- SERIALIZATION --


decoder : Decoder Build
decoder =
    decode Build
        |> required "id" (Decode.map Id int)
        |> required "status" statusDecoder
        |> required "commit" Commit.decodeHash
        |> required "project" Project.decodeId
        |> required "taskName" Task.decodeName
