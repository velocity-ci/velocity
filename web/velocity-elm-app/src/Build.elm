module Build exposing (Build, decoder)

import Build.Id as Id exposing (Id)
import Build.Status as Status exposing (Status)
import Iso8601
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, required)
import Json.Encode as Encode
import Time


type Build
    = Build Internals


type alias Internals =
    { id : Id
    , status : Status

    --    , task : Task
    --    , steps : List BuildStep
    , createdAt : Time.Posix
    , completedAt : Maybe Time.Posix
    , updatedAt : Maybe Time.Posix
    , startedAt : Maybe Time.Posix
    }



-- SERIALIZATION --


decoder : Decoder Build
decoder =
    Decode.succeed Build
        |> custom internalsDecoder


internalsDecoder : Decoder Internals
internalsDecoder =
    Decode.succeed Internals
        |> required "id" Id.decoder
        |> required "status" Status.decoder
        |> required "createdAt" Iso8601.decoder
        |> required "completedAt" Iso8601.decoder
        |> required "updatedAt" Iso8601.decoder
        |> required "startedAt" Iso8601.decoder



--                        decode Build
--                            |> required "id" (Decode.map Id string)
--                            |> required "status" statusDecoder
--                            |> required "task" Task.decoder
--                            |> required "buildSteps" (Decode.list BuildStep.decoder)
--                            |> required "createdAt" stringToDateTime
--                            |> required "completedAt" (Decode.maybe stringToDateTime)
--                            |> required "updatedAt" (Decode.maybe stringToDateTime)
--                            |> required "startedAt" (Decode.maybe stringToDateTime)
