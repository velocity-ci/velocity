module Page.Users exposing (..)

-- EXTERNAL --

import Http
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Task exposing (Task)
import Bootstrap.Modal as Modal
import Bootstrap.Button as Button
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
    , deletingUsers : List Username
    , form : UserForm.Context
    , formModalVisibility : Modal.Visibility
    , deleteModalVisibility : ( Maybe Username, Modal.Visibility )
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
    , deleteModalVisibility = ( Nothing, Modal.hidden )
    , deletingUsers = []
    }



-- SUBSCRIPTIONS --


subscriptions : Model -> Sub Msg
subscriptions { formModalVisibility, deleteModalVisibility } =
    Sub.batch
        [ Modal.subscriptions formModalVisibility AnimateFormModal
        , Modal.subscriptions (Tuple.second deleteModalVisibility) AnimateDeleteModal
        ]



-- VIEW --


view : Session msg -> Model -> Html Msg
view session model =
    div [ class "p-4 my-4" ]
        [ viewToolbar
        , viewIf (not (List.isEmpty model.users)) (viewUserList model.users model.deletingUsers)
        , viewFormModal model.form model.formModalVisibility
        , viewDeleteModal model.deleteModalVisibility
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


viewUserList : List Username -> List Username -> Html Msg
viewUserList users deletingUsers =
    let
        userLis =
            users
                |> List.sortBy User.usernameToString
                |> List.map (\u -> viewUserListItem u (List.member u deletingUsers))
    in
        div []
            [ h6 [] [ text "Users" ]
            , ul [ class "list-group list-group-flush" ] userLis
            ]


viewUserListItem : Username -> Bool -> Html Msg
viewUserListItem username deleting =
    li
        [ class "list-group-item align-items-start px-0 d-flex"
        , classList [ "text-muted" => deleting ]
        ]
        [ h6 [ class "w-100 align-self-center" ] [ text (User.usernameToString username) ]
        , div [ class "flex-shrink-1 align-self-bottom" ]
            [ Util.viewIf (not deleting)
                (Button.button
                    [ Button.outlineDanger
                    , Button.small
                    , Button.attrs
                        [ onClick (ShowDeleteModal username) ]
                    ]
                    [ i [ class "fa fa-trash" ] []
                    ]
                )
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


viewDeleteModal : ( Maybe Username, Modal.Visibility ) -> Html Msg
viewDeleteModal ( maybeUsername, visibility ) =
    case maybeUsername of
        Just username ->
            let
                header =
                    "Delete user '" ++ User.usernameToString username ++ "'"

                confirmButton =
                    Button.button
                        [ Button.outlinePrimary
                        , Button.attrs
                            [ onClick (DeleteUser username) ]
                        ]
                        [ text "Confirm" ]

                cancelButton =
                    Button.button
                        [ Button.outlineSecondary
                        , Button.attrs
                            [ onClick CloseDeleteModal ]
                        ]
                        [ text "Cancel" ]
            in
                Modal.config CloseDeleteModal
                    |> Modal.withAnimation AnimateDeleteModal
                    |> Modal.large
                    |> Modal.hideOnBackdropClick True
                    |> Modal.h3 [] [ text header ]
                    |> Modal.footer [] [ cancelButton, confirmButton ]
                    |> Modal.view visibility

        Nothing ->
            text ""


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
    | CloseDeleteModal
    | ShowFormModal
    | ShowDeleteModal Username
    | AnimateFormModal Modal.Visibility
    | AnimateDeleteModal Modal.Visibility
    | SetUsername String
    | SetPassword String
    | SetPasswordConfirm String
    | UserCreated (Result Request.Errors.HttpError Username)
    | SubmitForm
    | DeleteUser Username
    | UserDeleted (Result ( Request.Errors.HttpError, Username ) Username)


type ExternalMsg
    = NoOp
    | HandleRequestError Request.Errors.HandledError


update : Context -> Session msg -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update context session msg model =
    case msg of
        CloseDeleteModal ->
            { model | deleteModalVisibility = ( Nothing, Modal.hidden ) }
                => Cmd.none
                => NoOp

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

        AnimateDeleteModal visibility ->
            let
                ( maybeUsername, modalVisibility ) =
                    model.deleteModalVisibility
            in
                { model | deleteModalVisibility = ( maybeUsername, visibility ) }
                    => Cmd.none
                    => NoOp

        ShowDeleteModal username ->
            { model | deleteModalVisibility = ( Just username, Modal.shown ) }
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

        DeleteUser username ->
            let
                cmdFromAuth authToken =
                    username
                        |> Request.User.delete context (Just authToken)
                        |> Task.andThen (always <| Task.succeed username)
                        |> Task.mapError (\e -> ( e, username ))
                        |> Task.attempt UserDeleted

                cmd =
                    session
                        |> Session.attempt "delete user" cmdFromAuth
                        |> Tuple.second
            in
                { model
                    | deleteModalVisibility = ( Nothing, Modal.hidden )
                    , deletingUsers = username :: model.deletingUsers
                }
                    => cmd
                    => NoOp

        UserDeleted (Ok username) ->
            { model
                | users = List.filter (\a -> a /= username) model.users
                , deletingUsers = List.filter (\a -> a /= username) model.deletingUsers
            }
                => Cmd.none
                => NoOp

        UserDeleted (Err ( err, username )) ->
            { model | deletingUsers = List.filter (\a -> a /= username) model.deletingUsers }
                => Cmd.none
                => NoOp

        UserCreated (Ok username) ->
            { model
                | formModalVisibility = Modal.hidden
                , form = UserForm.init
                , users = username :: model.users
            }
                => Cmd.none
                => NoOp

        UserCreated (Err err) ->
            let
                globalError =
                    [ "" => "Unable to process user." ]

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
                                |> Form.updateServerErrors globalError UserForm.serverErrorToFormError
                                => NoOp
            in
                { model | form = Form.submitting False updatedForm }
                    => Cmd.none
                    => externalMsg
