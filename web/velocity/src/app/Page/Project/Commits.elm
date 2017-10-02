module Page.Project.Commits exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Data.Commit as Commit exposing (Commit)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Helpers exposing (formatDate, formatTime, sortByDatetime)
import Request.Project
import Util exposing ((=>))
import Task exposing (Task)
import Views.Page as Page
import Http
import Route exposing (Route)
import Page.Project.Route as ProjectRoute
import Dict exposing (Dict)
import Time.DateTime as DateTime exposing (DateTime)
import Time.Date as Date exposing (Date)
import Page.Helpers exposing (formatDate)


-- MODEL --


type alias Model =
    { commits : List Commit
    , submitting : Bool
    }


init : Session -> Project.Id -> Task PageLoadError Model
init session id =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadCommits =
            maybeAuthToken
                |> Request.Project.commits id
                |> Http.toTask

        initialModel commits =
            { commits = commits
            , submitting = False
            }

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."
    in
        Task.map initialModel loadCommits
            |> Task.mapError handleLoadError



-- VIEW --


commitListToDict : List Commit -> Dict ( Int, Int, Int ) (List Commit)
commitListToDict commits =
    let
        reducer commit dict =
            let
                date =
                    commit.date
                        |> DateTime.date
                        |> Date.toTuple

                insert =
                    case Dict.get date dict of
                        Just exists ->
                            commit :: exists

                        Nothing ->
                            [ commit ]
            in
                Dict.insert date insert dict
    in
        List.foldl reducer Dict.empty commits


view : Project -> Model -> Html Msg
view project model =
    commitListToDict model.commits
        |> viewCommitListContainer project


viewCommitListContainer : Project -> Dict ( Int, Int, Int ) (List Commit) -> Html Msg
viewCommitListContainer project dict =
    let
        listItemToDate dateListItem =
            dateListItem
                |> Tuple.first
                |> Date.fromTuple

        sortDateList a b =
            listItemToDate a
                |> Date.compare (listItemToDate b)
    in
        dict
            |> Dict.toList
            |> List.sortWith sortDateList
            |> List.map (viewCommitList project)
            |> div []


viewCommitList : Project -> ( ( Int, Int, Int ), List Commit ) -> Html Msg
viewCommitList project ( dateTuple, commits ) =
    let
        commitListItems =
            sortByDatetime .date commits
                |> List.map (viewCommitListItem project.id)

        formattedDate =
            Date.fromTuple dateTuple
                |> formatDate
    in
        div [ class "default-margin-bottom" ]
            [ h6 [ class "mb-2 text-muted" ] [ text formattedDate ]
            , div [ class "card" ]
                [ div [ class "list-group list-group-flush" ] commitListItems
                ]
            ]


viewCommitListItem : Project.Id -> Commit -> Html Msg
viewCommitListItem id commit =
    let
        truncatedHash =
            Commit.truncateHash commit.hash

        route =
            Route.Project (ProjectRoute.Commit commit.hash) id
    in
        a [ class "list-group-item list-group-item-action flex-column align-items-start", Route.href route ]
            [ div [ class "d-flex w-100 justify-content-between" ]
                [ h5 [ class "mb-1 text-overflow" ] [ text commit.message ]
                , small [] [ text truncatedHash ]
                ]
            , small [] [ strong [] [ text commit.author ], text (" commited at " ++ formatTime commit.date) ]
            ]


breadcrumb : Project -> List ( Route, String )
breadcrumb project =
    [ ( Route.Project ProjectRoute.Commits project.id, "Commits" ) ]


viewBreadcrumbExtraItems : Model -> Html Msg
viewBreadcrumbExtraItems model =
    div [ class "ml-auto p-2" ]
        [ button
            [ class "ml-auto btn btn-dark", type_ "button", onClick SubmitSync, disabled model.submitting ]
            [ i [ class "fa fa-refresh" ] [], text " Refresh " ]
        ]



-- UPDATE --


type Msg
    = SubmitSync
    | SyncCompleted (Result Http.Error (List Commit))


update : Project -> Session -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    case msg of
        SubmitSync ->
            let
                getCommits authToken =
                    Request.Project.commits project.id (Just authToken)
                        |> Http.toTask

                cmdFromAuth authToken =
                    authToken
                        |> Request.Project.sync project.id
                        |> Http.toTask
                        |> Task.andThen (getCommits authToken |> always)
                        |> Task.attempt SyncCompleted

                cmd =
                    session
                        |> Session.attempt "sync project" cmdFromAuth
                        |> Tuple.second
            in
                { model | submitting = True } => cmd

        SyncCompleted (Ok commits) ->
            { model
                | submitting = False
                , commits = commits
            }
                => Cmd.none

        SyncCompleted (Err err) ->
            { model | submitting = False } => Cmd.none
