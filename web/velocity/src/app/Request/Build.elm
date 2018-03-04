module Request.Build exposing (..)

import Data.Build as Build exposing (Build)
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.BuildStream as BuildStream exposing (BuildStream, BuildStreamOutput)
import Data.PaginatedList as PaginatedList exposing (PaginatedList)
import Data.AuthToken as AuthToken exposing (AuthToken, withAuthorization)
import Request.Helpers exposing (apiUrl)
import Request.Errors
import HttpBuilder exposing (RequestBuilder, withBody, withExpect, withQueryParams)
import Http
import Array exposing (Array)
import Json.Decode as Decode
import Task exposing (Task)


steps :
    Build.Id
    -> Maybe AuthToken
    -> Task Request.Errors.HttpError (PaginatedList BuildStep)
steps id maybeToken =
    let
        expect =
            BuildStep.decoder
                |> PaginatedList.decoder
                |> Http.expectJson
    in
        apiUrl ("/builds" ++ "/" ++ Build.idToString id ++ "/steps")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleError


streams :
    Maybe AuthToken
    -> BuildStep.Id
    -> Task Request.Errors.HttpError (PaginatedList BuildStream)
streams maybeToken id =
    let
        expect =
            BuildStream.decoder
                |> PaginatedList.decoder
                |> Http.expectJson
    in
        apiUrl ("/steps" ++ "/" ++ BuildStep.idToString id ++ "/streams")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleError


streamOutput :
    Maybe AuthToken
    -> BuildStream.Id
    -> Task Request.Errors.HttpError (Array BuildStreamOutput)
streamOutput maybeToken id =
    let
        expect =
            BuildStream.outputDecoder
                |> Decode.array
                |> Decode.at [ "data" ]
                |> Http.expectJson
    in
        apiUrl ("/streams/" ++ BuildStream.idToString id ++ "/log")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleError
