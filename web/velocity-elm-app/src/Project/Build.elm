module Project.Build exposing (Build, decoder)

import Iso8601
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, required)
import Json.Encode as Encode
import Project.Build.Id as Id exposing (Id)
import Project.Build.Status as Status exposing (Status)
import Project.Build.Step as Step exposing (Step)
import Project.Task as Task exposing (Task)
import Time


type Build
    = Build Internals


type alias Internals =
    { id : Id
    , status :
        Status
        --    , task : Task
    , steps : List Step
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
        |> required "buildSteps" (Decode.list Step.decoder)
        |> required "createdAt" Iso8601.decoder
        |> required "completedAt" (Decode.maybe Iso8601.decoder)
        |> required "updatedAt" (Decode.maybe Iso8601.decoder)
        |> required "startedAt" (Decode.maybe Iso8601.decoder)



--                        decode Build
--                            |> required "id" (Decode.map Id string)
--                            |> required "status" statusDecoder
--                            |> required "task" Task.decoder
--                            |> required "buildSteps" (Decode.list BuildStep.decoder)
--                            |> required "createdAt" stringToDateTime
--                            |> required "completedAt" (Decode.maybe stringToDateTime)
--                            |> required "updatedAt" (Decode.maybe stringToDateTime)
--                            |> required "startedAt" (Decode.maybe stringToDateTime)
