module Page.Users exposing (..)

-- EXTERNAL --

import Http
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Task exposing (Task)
import Bootstrap.Modal as Modal
import Json.Decode exposing (decodeString)


-- INTERNAL --

import Context exposing (Context)
import Data.User as User exposing (Username)
import Data.Session as Session exposing (Session)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.Errors
import Request.User
import Views.Page as Page
import Util exposing ((=>), viewIf)
import Component.Form as Form
import Component.UserForm as UserForm


-- MODEL --


type alias Model =
    { users : List Username
    , form : UserForm.Context
    , formModalVisibility : Modal.Visibility
    }


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
        Task.map initialModel users
            |> Task.mapError (Request.Errors.withDefaultError loadError)


initialModel : List Username -> Model
initialModel users =
    { users = users
    , form = UserForm.init
    , formModalVisibility = Modal.hidden
    }



-- SUBSCRIPTIONS --


subscriptions : Model -> Sub Msg
subscriptions { formModalVisibility } =
    Modal.subscriptions formModalVisibility AnimateFormModal



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
            , viewFormModal model.form model.formModalVisibility
            ]


viewToolbar : Html Msg
viewToolbar =
    div [ class "btn-toolbar d-flex flex-row-reverse" ]
        [ button
            [ onClick ShowFormModal
            , class "btn btn-primary btn-lg"
            , style [ "border-radius" => "25px" ]
            ]
            [ i [ class "fa fa-plus" ] [] ]
        ]


viewUserList : List Username -> Html Msg
viewUserList users =
    let
        userLis =
            users
                |> List.sortBy User.usernameToString
                |> List.map viewUserListItem
    in
        div []
            [ h6 [] [ text "Users" ]
            , ul [ class "list-group list-group-flush" ] userLis
            ]


viewUserListItem : Username -> Html Msg
viewUserListItem username =
    li [ class "list-group-item align-items-start px-0" ]
        [ div [ class "d-flex w-100 justify-content-between" ]
            [ h6 [ class "mb-1" ] [ text (User.usernameToString username) ]
            ]
        ]


viewFormModal : UserForm.Context -> Modal.Visibility -> Html Msg
viewFormModal userForm visibility =
    Modal.config CloseFormModal
        |> Modal.withAnimation AnimateFormModal
        |> Modal.large
        |> Modal.hideOnBackdropClick True
        |> Modal.h3 [] [ text "Add user" ]
        |> Modal.body [] [ UserForm.view userFormConfig userForm ]
        |> Modal.footer [] [ UserForm.viewSubmitButton userFormConfig userForm ]
        |> Modal.view visibility


userFormConfig : UserForm.Config Msg
userFormConfig =
    { setUsernameMsg = SetUsername
    , setPasswordMsg = SetPassword
    , setPasswordConfirmMsg = SetPasswordConfirm
    , submitMsg = SubmitForm
    }



-- UPDATE --


type Msg
    = CloseFormModal
    | ShowFormModal
    | AnimateFormModal Modal.Visibility
    | SetUsername String
    | SetPassword String
    | SetPasswordConfirm String
    | UserCreated (Result Request.Errors.HttpError Username)
    | SubmitForm


type ExternalMsg
    = NoOp
    | HandleRequestError Request.Errors.HandledError


update : Context -> Session msg -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update context session msg model =
    case msg of
        CloseFormModal ->
            { model
                | formModalVisibility = Modal.hidden
                , form = UserForm.init
            }
                => Cmd.none
                => NoOp

        AnimateFormModal visibility ->
            { model | formModalVisibility = visibility }
                => Cmd.none
                => NoOp

        ShowFormModal ->
            { model | formModalVisibility = Modal.shown }
                => Cmd.none
                => NoOp

        SetUsername value ->
            { model | form = UserForm.update model.form UserForm.username value }
                => Cmd.none
                => NoOp

        SetPassword value ->
            { model | form = UserForm.update model.form UserForm.password value }
                => Cmd.none
                => NoOp

        SetPasswordConfirm value ->
            { model | form = UserForm.update model.form UserForm.passwordConfirm value }
                => Cmd.none
                => NoOp

        SubmitForm ->
            let
                cmdFromAuth authToken =
                    model.form
                        |> UserForm.submitValues
                        |> Request.User.create context (Just authToken)
                        |> Task.attempt UserCreated

                cmd =
                    session
                        |> Session.attempt "create user" cmdFromAuth
                        |> Tuple.second
            in
                { model | form = Form.submit model.form }
                    => cmd
                    => NoOp

        UserCreated (Ok user) ->
            { model
                | formModalVisibility = Modal.hidden
                , form = UserForm.init
                , users = user :: model.users
            }
                => Cmd.none
                => NoOp

        UserCreated (Err err) ->
            let
                ( updatedForm, externalMsg ) =
                    case err of
                        Request.Errors.HandledError handledError ->
                            model.form
                                => HandleRequestError handledError

                        Request.Errors.UnhandledError (Http.BadStatus response) ->
                            let
                                errors =
                                    response.body
                                        |> decodeString UserForm.errorsDecoder
                                        |> Result.withDefault []
                            in
                                model.form
                                    |> Form.updateServerErrors errors UserForm.serverErrorToFormError
                                    => NoOp

                        _ ->
                            model.form
                                |> Form.updateServerErrors [ "" => "Unable to process user." ] UserForm.serverErrorToFormError
                                => NoOp
            in
                { model | form = Form.submitting False updatedForm }
                    => Cmd.none
                    => externalMsg
