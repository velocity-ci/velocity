module Request.Project
    exposing
        ( list
        , create
        , get
        , sync
        , delete
        , branches
        , builds
        )

import Data.AuthToken as AuthToken exposing (AuthToken, withAuthorization)
import Data.Project as Project exposing (Project)
import Data.Branch as Branch exposing (Branch)
import Data.PaginatedList as PaginatedList exposing (PaginatedList)
import Data.Build as Build exposing (Build)
import Json.Encode as Encode
import Request.Helpers exposing (apiUrl)
import Request.Errors
import HttpBuilder exposing (RequestBuilder, withBody, withExpect, withQueryParams)
import Util exposing ((=>))
import Http
import Task exposing (Task)


baseUrl : String
baseUrl =
    "/projects"



-- LIST --


list : Maybe AuthToken -> Task Request.Errors.HttpError (PaginatedList Project)
list maybeToken =
    let
        expect =
            Project.decoder
                |> PaginatedList.decoder
                |> Http.expectJson
    in
        apiUrl baseUrl
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleError



-- SYNC --


sync : Project.Slug -> AuthToken -> Task Request.Errors.HttpError Project
sync slug token =
    let
        expect =
            Project.decoder
                |> Http.expectJson
    in
        apiUrl (baseUrl ++ "/" ++ Project.slugToString slug ++ "/sync")
            |> HttpBuilder.post
            |> withAuthorization (Just token)
            |> withExpect expect
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleError



-- BRANCHES --


branches : Project.Slug -> Maybe AuthToken -> Task Request.Errors.HttpError (PaginatedList Branch)
branches slug maybeToken =
    let
        expect =
            Branch.decoder
                |> PaginatedList.decoder
                |> Http.expectJson
    in
        apiUrl (baseUrl ++ "/" ++ Project.slugToString slug ++ "/branches")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleError



-- BUILDS --


builds : Project.Slug -> Maybe AuthToken -> Task Request.Errors.HttpError (PaginatedList Build)
builds slug maybeToken =
    let
        expect =
            Build.decoder
                |> PaginatedList.decoder
                |> Http.expectJson
    in
        apiUrl (baseUrl ++ "/" ++ Project.slugToString slug ++ "/builds")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleError



-- GET --


get : Project.Slug -> Maybe AuthToken -> Task Request.Errors.HttpError Project
get slug maybeToken =
    let
        expect =
            Project.decoder
                |> Http.expectJson
    in
        apiUrl (baseUrl ++ "/" ++ Project.slugToString slug)
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleError



-- CREATE --


type alias CreateConfig record =
    { record
        | name : String
        , repository : String
        , privateKey : Maybe String
    }


create : CreateConfig record -> AuthToken -> Task Request.Errors.HttpError Project
create config token =
    let
        expect =
            Project.decoder
                |> Http.expectJson

        baseProject =
            [ "name" => Encode.string config.name
            , "address" => Encode.string config.repository
            ]

        project =
            config.privateKey
                |> Maybe.map (\privateKey -> ( "key", Encode.string privateKey ) :: baseProject)
                |> Maybe.withDefault baseProject

        body =
            project
                |> Encode.object
                |> Http.jsonBody
    in
        apiUrl baseUrl
            |> HttpBuilder.post
            |> withAuthorization (Just token)
            |> withBody body
            |> withExpect expect
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleError



-- DELETE --


delete : Project.Slug -> AuthToken -> Task Request.Errors.HttpError ()
delete slug token =
    apiUrl (baseUrl ++ "/" ++ Project.slugToString slug)
        |> HttpBuilder.delete
        |> withAuthorization (Just token)
        |> withExpect (Http.expectStringResponse (\_ -> Ok ()))
        |> HttpBuilder.toTask
        |> Task.mapError Request.Errors.handleError
