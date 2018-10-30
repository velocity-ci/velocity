module Request.Build exposing (steps, streamOutput, streams)

import Array exposing (Array)
import Context exposing (Context)
import Data.AuthToken as AuthToken exposing (AuthToken, withAuthorization)
import Data.Build as Build exposing (Build)
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.BuildStream as BuildStream exposing (BuildStream, BuildStreamOutput)
import Data.PaginatedList as PaginatedList exposing (PaginatedList)
import Http
import HttpBuilder exposing (RequestBuilder, withBody, withExpect, withQueryParams)
import Json.Decode as Decode
import Request.Errors
import Request.Helpers exposing (apiUrl)
import Task exposing (Task)


steps :
    Context
    -> Build.Id
    -> Maybe AuthToken
    -> Task Request.Errors.HttpError (PaginatedList BuildStep)
steps context id maybeToken =
    let
        expect =
            BuildStep.decoder
                |> PaginatedList.decoder
                |> Http.expectJson
    in
        apiUrl context ("/builds" ++ "/" ++ Build.idToString id ++ "/steps?amount=-1")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleHttpError


streams :
    Context
    -> Maybe AuthToken
    -> BuildStep.Id
    -> Task Request.Errors.HttpError (PaginatedList BuildStream)
streams context maybeToken id =
    let
        expect =
            BuildStream.decoder
                |> PaginatedList.decoder
                |> Http.expectJson
    in
        apiUrl context ("/steps" ++ "/" ++ BuildStep.idToString id ++ "/streams?amount=-1")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleHttpError


streamOutput :
    Context
    -> Maybe AuthToken
    -> BuildStream.Id
    -> Task Request.Errors.HttpError (Array BuildStreamOutput)
streamOutput context maybeToken id =
    let
        expect =
            BuildStream.outputDecoder
                |> Decode.array
                |> Decode.at [ "data" ]
                |> Http.expectJson
    in
        apiUrl context ("/streams/" ++ BuildStream.idToString id ++ "/log?amount=-1")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleHttpError
