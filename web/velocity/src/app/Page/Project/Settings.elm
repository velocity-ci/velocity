module Page.Project.Settings exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Util exposing ((=>))
import Route exposing (Route)
import Page.Project.Route as ProjectRoute
import Request.Project
import Http
import Route


-- MODEL --


type alias Model =
    { deleting : Bool
    , deleteConfirmationToggled : Bool
    , deleteConfirmProjectName : String
    }


initialModel : Model
initialModel =
    { deleting = False
    , deleteConfirmationToggled = False
    , deleteConfirmProjectName = ""
    }



-- VIEW --


view : Project -> Model -> Html Msg
view project model =
    div []
        [ viewDangerArea project model ]


viewDangerArea : Project -> Model -> Html Msg
viewDangerArea project model =
    let
        cardBody =
            if model.deleteConfirmationToggled then
                viewDeleteConfirmation project.name model.deleteConfirmProjectName
            else
                viewPreDeleteConfirmation model
    in
        div [ class "card border-danger mb-3" ]
            [ div [ class "card-body" ]
                [ h5 [ class "card-title" ] [ text "Delete project" ]
                , cardBody
                ]
            ]


viewPreDeleteConfirmation : Model -> Html Msg
viewPreDeleteConfirmation model =
    div []
        [ p
            [ class "card-text" ]
            [ text "Once you delete a project, there is no going back. Please be certain." ]
        , button
            [ type_ "button"
            , class "btn btn-outline-danger"
            , onClick (ToggleDeleteConfirmation True)
            , disabled model.deleting
            ]
            [ text "Delete project" ]
        ]


viewDeleteConfirmation : String -> String -> Html Msg
viewDeleteConfirmation projectName confirmValue =
    let
        disclaimer =
            div []
                [ p []
                    [ text "This will permanently delete the "
                    , strong [] [ text projectName ]
                    , text " project and builds."
                    ]
                , p []
                    [ text "Please type in the name of the project to confirm or "
                    , button
                        [ type_ "button"
                        , class "btn btn-link btn-cancel-delete"
                        , onClick (ToggleDeleteConfirmation False)
                        ]
                        [ text "click here to cancel." ]
                    ]
                ]
    in
        div []
            [ disclaimer
            , div [ class "input-group" ]
                [ input
                    [ class "form-control"
                    , type_ "text"
                    , value confirmValue
                    , onInput SetConfirmProjectName
                    ]
                    []
                , span [ class "input-group-btn" ]
                    [ button
                        [ class "btn btn-danger"
                        , type_ "button"
                        , disabled (projectName /= confirmValue)
                        , onClick SubmitProjectDelete
                        ]
                        [ text "Delete project" ]
                    ]
                ]
            ]


breadcrumb : Project -> List ( Route, String )
breadcrumb project =
    [ ( Route.Project ProjectRoute.Settings project.id, "Settings" ) ]



-- UPDATE --


type Msg
    = ToggleDeleteConfirmation Bool
    | SubmitProjectDelete
    | ProjectDeleted (Result Http.Error ())
    | SetConfirmProjectName String


update : Project -> Session -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    case msg of
        ToggleDeleteConfirmation state ->
            { model
                | deleteConfirmationToggled = state
                , deleteConfirmProjectName = ""
            }
                => Cmd.none

        SetConfirmProjectName name ->
            { model | deleteConfirmProjectName = name }
                => Cmd.none

        SubmitProjectDelete ->
            let
                cmdFromAuth authToken =
                    authToken
                        |> Request.Project.delete project.id
                        |> Http.send ProjectDeleted

                cmd =
                    session
                        |> Session.attempt "delete project" cmdFromAuth
                        |> Tuple.second
            in
                { model | deleting = True }
                    => cmd

        ProjectDeleted (Ok _) ->
            { model | deleting = False }
                => Route.modifyUrl Route.Projects

        ProjectDeleted (Err _) ->
            { model | deleting = False }
                => Cmd.none
