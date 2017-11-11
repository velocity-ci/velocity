module Data.Build exposing (..)

import Data.Project as Project
import Data.Commit as Commit
import Data.Task as Task
import Json.Decode as Decode exposing (Decoder, int, string)
import Json.Decode.Pipeline as Pipeline exposing (custom, decode, hardcoded, required, optional)


type alias Build =
    { id : Id
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


statusDecoder : Decoder Status
statusDecoder =
    Decode.string
        |> Decode.andThen
            (\status ->
                case status of
                    "waiting" ->
                        Decode.succeed Waiting

                    "failed" ->
                        Decode.succeed Failed

                    "running" ->
                        Decode.succeed Running

                    unknown ->
                        Decode.fail <| "Unknown status: " ++ unknown
            )



-- IDENTIFIERS --


type Status
    = Waiting
    | Failed
    | Running


type Id
    = Id Int


idToString : Id -> String
idToString (Id id) =
    toString id
