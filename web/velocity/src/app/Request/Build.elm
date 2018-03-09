module Request.Build exposing (..)

import Context exposing (Context)
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
        apiUrl context ("/builds" ++ "/" ++ Build.idToString id ++ "/steps")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleError


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
        apiUrl context ("/steps" ++ "/" ++ BuildStep.idToString id ++ "/streams")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleError


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
        apiUrl context ("/streams/" ++ BuildStream.idToString id ++ "/log")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleError
