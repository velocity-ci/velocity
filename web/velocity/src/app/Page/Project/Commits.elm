module Page.Project.Commits exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Data.Commit as Commit exposing (Commit)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Helpers exposing (formatDate)
import Request.Project
import Util exposing ((=>))
import Task exposing (Task)
import Views.Page as Page
import Http


-- MODEL --


type alias Model =
    { commits : List Commit }


init : Session -> Project.Id -> Task PageLoadError Model
init session id =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadCommits =
            maybeAuthToken
                |> Request.Project.commits id
                |> Http.toTask

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."
    in
        Task.map Model loadCommits
            |> Task.mapError handleLoadError



-- VIEW --


view : Model -> Html Msg
view model =
    div []
        [ viewCommitList model.commits ]


viewCommitList : List Commit -> Html Msg
viewCommitList commits =
    List.map viewCommitListItem commits
        |> div []


viewCommitListItem : Commit -> Html Msg
viewCommitListItem commit =
    let
        authorAndDate =
            [ strong [] [ text commit.author ], text (" commited on " ++ formatDate commit.date) ]

        truncatedHash =
            String.slice 0 7 commit.hash
    in
        div [ class "list-group" ]
            [ a [ class "list-group-item list-group-item-action flex-column align-items-start", href "#" ]
                [ div [ class "d-flex w-100 justify-content-between" ]
                    [ h5 [ class "mb-1" ] [ text commit.message ]
                    , small [] [ text truncatedHash ]
                    ]
                , small [] authorAndDate
                ]
            ]



-- UPDATE --


type Msg
    = NoOp


update : Session -> Msg -> Model -> ( Model, Cmd Msg )
update session msg model =
    model => Cmd.none
