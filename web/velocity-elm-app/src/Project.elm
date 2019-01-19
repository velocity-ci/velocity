module Project
    exposing
        ( Hydrated
        , Project
        , addProject
        , create
        , findProjectById
        , findProjectBySlug
        , id
        , list
        , name
        , repository
        , slug
        , syncing
        , thumbnail
        , updateProject
        , CreateResponse(..)
        , selectionSet
        )

import Api exposing (BaseUrl, Cred)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Font as Font
import Icon
import Iso8601
import Palette
import Project.Branch as Branch exposing (Branch)
import Project.Id as Id exposing (Id)
import Project.Slug as Slug exposing (Slug)
import Time exposing (Posix)
import Graphql.Http
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, hardcoded, with)
import Api.Compiled.Object.Project as Project
import Api.Compiled.Object
import Api.Compiled.Scalar
import Api.Compiled.Query as Query
import Api.Compiled.Object.ProjectPayload as ProjectPayload
import Api.Compiled.Mutation as Mutation


type alias Hydrated =
    { project : Project
    , branches : List Branch
    }


type Project
    = Project Internals


type alias Internals =
    { id : Id
    , slug : Slug
    , name : String
    , address :
        String
        --    , createdAt : Time.Posix
        --    , updatedAt : Time.Posix
    , synchronising : Bool
    , logo : Maybe String
    , branches : List Branch
    }



-- SERIALIZATION --


selectionSet : SelectionSet Project Api.Compiled.Object.Project
selectionSet =
    SelectionSet.succeed Project
        |> with internalSelectionSet


internalSelectionSet : SelectionSet Internals Api.Compiled.Object.Project
internalSelectionSet =
    SelectionSet.succeed Internals
        |> with Id.selectionSet
        |> with Slug.selectionSet
        |> with Project.name
        |> with Project.address
        --        |> with (mapToDateTime Project.insertedAt)
        --        |> with (mapToDateTime Project.updatedAt)
        |>
            hardcoded False
        |> hardcoded Nothing
        |> with (Project.branches Branch.selectionSet)


mapToDateTime : SelectionSet Api.Compiled.Scalar.NaiveDateTime typeLock -> SelectionSet Posix typeLock
mapToDateTime =
    SelectionSet.mapOrFail
        (\(Api.Compiled.Scalar.NaiveDateTime value) ->
            Iso8601.toTime value
                |> Result.mapError
                    (\_ ->
                        (Debug.log "decode time error"
                            ("Failed to parse "
                                ++ value
                                ++ " as Iso8601 DateTime."
                            )
                        )
                    )
        )



-- INFO --


id : Project -> Id
id (Project project) =
    project.id


name : Project -> String
name (Project project) =
    project.name


slug : Project -> Slug
slug (Project project) =
    project.slug


repository : Project -> String
repository (Project project) =
    project.address


syncing : Project -> Bool
syncing (Project project) =
    project.synchronising



-- ELEMENTS --


thumbnail : Project -> Element msg
thumbnail project =
    case thumbnailSrc project of
        Just src ->
            el
                [ centerX
                , centerY
                , width fill
                , height fill
                , Background.uncropped src
                , Border.width 1
                , Border.color Palette.neutral5
                , Border.rounded 10
                , padding 5
                ]
                (text "")

        Nothing ->
            el
                [ height fill
                , width fill
                , Border.width 1
                , Border.color Palette.neutral5
                , Border.rounded 10
                , paddingXY 5 0
                , Font.color Palette.neutral6
                ]
                (el [ centerX, centerY ] <| Icon.code Icon.fullSizeOptions)


thumbnailSrc : Project -> Maybe String
thumbnailSrc (Project project) =
    project.logo



-- HELPERS --


findProjectById : List Project -> Id -> Maybe Project
findProjectById projects targetId =
    List.filter (\(Project b) -> b.id == targetId) projects
        |> List.head


findProjectBySlug : List Project -> Slug -> Maybe Project
findProjectBySlug projects targetSlug =
    List.filter (\(Project b) -> b.slug == targetSlug) projects
        |> List.head


addProject : Project -> List Project -> List Project
addProject (Project internals) projects =
    case findProjectById projects internals.id of
        Just _ ->
            updateProject (Project internals) projects

        Nothing ->
            Project internals :: projects


updateProject : Project -> List Project -> List Project
updateProject (Project a) projects =
    projects
        |> List.map
            (\(Project b) ->
                if a.id == b.id then
                    Project a
                else
                    Project b
            )



-- COLLECTION --


list : Cred -> BaseUrl -> Graphql.Http.Request (List Project)
list cred baseUrl =
    selectionSet
        |> Query.listProjects
        |> Api.queryRequest baseUrl


type CreateResponse
    = CreateSuccess Project
    | ValidationFailure (List Api.ValidationMessage)


createResponseSelectionSet : SelectionSet CreateResponse Api.Compiled.Object.ProjectPayload
createResponseSelectionSet =
    let
        messageSelectionSet =
            ProjectPayload.messages Api.validationErrorSelectionSet
                |> SelectionSet.withDefault []
                |> SelectionSet.nonNullElementsOrFail

        toResponse messages result =
            case result of
                Just knownHost ->
                    CreateSuccess knownHost

                Nothing ->
                    ValidationFailure messages
    in
        SelectionSet.succeed toResponse
            |> SelectionSet.with messageSelectionSet
            |> SelectionSet.with (ProjectPayload.result selectionSet)


create : Cred -> BaseUrl -> Mutation.CreateProjectRequiredArguments -> Graphql.Http.Request CreateResponse
create cred baseUrl values =
    Mutation.createProject values createResponseSelectionSet
        |> Api.mutationRequest baseUrl
