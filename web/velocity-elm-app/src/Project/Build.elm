module Project.Build exposing (Build, decoder, list)

import Api exposing (BaseUrl, Cred)
import Api.Endpoint as Endpoint
import Http
import Iso8601
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, required)
import Json.Encode as Encode
import PaginatedList exposing (PaginatedList)
import Project.Build.Id as Id exposing (Id)
import Project.Build.Status as Status exposing (Status)
import Project.Build.Step as Step exposing (Step)
import Project.Slug as ProjectSlug
import Project.Task as Task exposing (Task)
import Time


type Build
    = Build Internals


type alias Internals =
    { id : Id
    , status : Status
    , task : Task
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
        |> required "task" Task.decoder
        |> required "buildSteps" (Decode.list Step.decoder)
        |> required "createdAt" Iso8601.decoder
        |> required "completedAt" (Decode.maybe Iso8601.decoder)
        |> required "updatedAt" (Decode.maybe Iso8601.decoder)
        |> required "startedAt" (Decode.maybe Iso8601.decoder)



-- COLLECTION --


list : Cred -> BaseUrl -> ProjectSlug.Slug -> Http.Request (PaginatedList Build)
list cred baseUrl projectSlug =
    let
        endpoint =
            Endpoint.builds (Just { amount = -1, page = 1 }) (Api.toEndpoint baseUrl) projectSlug
    in
    PaginatedList.decoder decoder
        |> Api.get endpoint (Just cred)
