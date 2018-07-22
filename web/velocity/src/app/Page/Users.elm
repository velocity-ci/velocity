module Page.Users exposing (..)

-- EXTERNAL --

import Html exposing (..)
import Html.Attributes exposing (..)
import Task exposing (Task)


-- INTERNAL --

import Context exposing (Context)
import Data.User as User exposing (Username)
import Data.Session exposing (Session)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.Errors
import Request.User
import Views.Page as Page
import Util exposing ((=>), viewIf)


-- MODEL --


type alias Model =
    { users : List Username }


init : Context -> Session msg -> Task (Request.Errors.Error PageLoadError) Model
init context session =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        users =
            Request.User.list context maybeAuthToken

        loadError =
            pageLoadError Page.Users "Users are currently unavailable."
    in
        Task.map Model users
            |> Task.mapError (Request.Errors.withDefaultError loadError)



-- VIEW --


view : Session msg -> Model -> Html Msg
view session model =
    let
        hasUsers =
            not (List.isEmpty model.users)

        userList =
            viewUserList model.users
    in
        div [ class "p-4 my-4" ]
            [ viewToolbar
            , viewIf hasUsers userList
              --            , viewFormModal model.form model.formModalVisibility
            ]


viewToolbar : Html Msg
viewToolbar =
    div [ class "btn-toolbar d-flex flex-row-reverse" ]
        [ button
            [ class "btn btn-primary btn-lg"
            , style [ "border-radius" => "25px" ]
            ]
            [ i [ class "fa fa-plus" ] [] ]
        ]


viewUserList : List Username -> Html Msg
viewUserList users =
    div []
        [ h6 [] [ text "Users" ]
        , ul [ class "list-group list-group-flush" ] (List.map viewUserListItem users)
        ]


viewUserListItem : Username -> Html Msg
viewUserListItem username =
    li [ class "list-group-item align-items-start px-0" ]
        [ div [ class "d-flex w-100 justify-content-between" ]
            [ h6 [ class "mb-1" ] [ text (User.usernameToString username) ]
            ]
        ]



-- UPDATE --


type Msg
    = NoOp_


type ExternalMsg
    = NoOp
    | HandleRequestError Request.Errors.HandledError


update : Context -> Session msg -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update context session msg model =
    ( ( model, Cmd.none ), NoOp )
