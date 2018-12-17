module Project exposing
    ( Hydrated
    , Project
    , addProject
    , channel
    , channelName
    , create
    , decoder
    , findProjectById
    , findProjectBySlug
    , id
    , list
    , name
    , repository
    , slug
    , sync
    , syncing
    , thumbnail
    , updateProject
    )

import Api exposing (BaseUrl, Cred)
import Api.Endpoint as Endpoint exposing (Endpoint)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Font as Font
import Http
import Icon
import Iso8601
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, required)
import Json.Encode as Encode
import PaginatedList exposing (PaginatedList)
import Palette
import Phoenix.Channel as Channel exposing (Channel)
import Project.Branch as Branch exposing (Branch)
import Project.Id as Id exposing (Id)
import Project.Slug as Slug exposing (Slug)
import Task exposing (Task)
import Time


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
    , repository : String
    , createdAt : Time.Posix
    , updatedAt : Time.Posix
    , synchronising : Bool
    , logo : Maybe String
    }



-- SERIALIZATION --


decoder : Decoder Project
decoder =
    Decode.succeed Project
        |> custom internalsDecoder


internalsDecoder : Decoder Internals
internalsDecoder =
    Decode.succeed Internals
        |> required "id" Id.decoder
        |> required "slug" Slug.decoder
        |> required "name" Decode.string
        |> required "repository" Decode.string
        |> required "createdAt" Iso8601.decoder
        |> required "updatedAt" Iso8601.decoder
        |> required "synchronising" Decode.bool
        |> required "logo" (Decode.maybe Decode.string)



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
    project.repository


channelName : Project -> String
channelName (Project project) =
    "project:" ++ Slug.toString project.slug


channel : Project -> Channel msg
channel project =
    Channel.init (channelName project)


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
            projects

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


list : Cred -> BaseUrl -> Task Http.Error (List Project)
list cred baseUrl =
    let
        endpoint =
            Endpoint.projects (Just { amount = -1, page = 1 }) (Api.toEndpoint baseUrl)
    in
    Decode.field "data" (Decode.list decoder)
        |> Api.getTask endpoint (Just cred)


sync : Cred -> BaseUrl -> Slug -> (Result Http.Error Project -> msg) -> Cmd msg
sync cred baseUrl slug_ toMsg =
    let
        endpoint =
            Endpoint.projectSync (Api.toEndpoint baseUrl) slug_
    in
    Api.post endpoint (Just cred) (Encode.object [] |> Http.jsonBody) toMsg decoder


create : Cred -> BaseUrl -> { a | name : String, repository : String, privateKey : Maybe String } -> (Result Http.Error Project -> msg) -> Cmd msg
create cred baseUrl values toMsg =
    let
        endpoint =
            Endpoint.projects Nothing (Api.toEndpoint baseUrl)

        baseValues =
            [ ( "name", Encode.string values.name )
            , ( "address", Encode.string values.repository )
            ]

        submitValues =
            case values.privateKey of
                Just privateKey ->
                    ( "key", Encode.string privateKey ) :: baseValues

                Nothing ->
                    baseValues

        body =
            submitValues
                |> Encode.object
                |> Http.jsonBody
    in
    Api.post endpoint (Just cred) body toMsg decoder



-- CHANNEL --
