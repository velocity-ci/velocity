module Page.KnownHosts exposing (..)

import Context exposing (Context)
import Data.KnownHost as KnownHost exposing (KnownHost)
import Data.Session as Session exposing (Session)
import Data.PaginatedList as PaginatedList exposing (Paginated(..))
import Task exposing (Task)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.KnownHost
import Request.Errors
import Views.Page as Page
import Http
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, onSubmit)
import Util exposing ((=>))
import Json.Decode as Decode exposing (Decoder, decodeString, field, string)
import Component.KnownHostForm as KnownHostForm
import Bootstrap.Modal as Modal
import Util exposing ((=>), onClickStopPropagation, viewIf)


-- MODEL --


type alias Model =
    { formModalVisibility : Modal.Visibility
    , form : KnownHostForm.Context
    , knownHosts : List KnownHost
    }


init : Context -> Session msg -> Task (Request.Errors.Error PageLoadError) Model
init context session =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadKnownHosts =
            Request.KnownHost.list context maybeAuthToken

        loadError =
            pageLoadError Page.KnownHosts "Known hosts are currently unavailable."

        initialModel (Paginated { total, results }) =
            { formModalVisibility = Modal.hidden
            , form = KnownHostForm.init
            , knownHosts = results
            }
    in
        Task.map initialModel loadKnownHosts
            |> Task.mapError (Request.Errors.withDefaultError loadError)



-- SUBSCRIPTIONS --


subscriptions : Model -> Sub Msg
subscriptions { formModalVisibility } =
    Modal.subscriptions formModalVisibility AnimateFormModal



-- VIEW --


view : Session msg -> Model -> Html Msg
view session model =
    let
        hasKnownHosts =
            not (List.isEmpty model.knownHosts)

        knownHostList =
            viewKnownHostList model.knownHosts
    in
        div [ class "py-2 my-4" ]
            [ viewToolbar
            , viewIf hasKnownHosts knownHostList
            , viewFormModal model.form model.formModalVisibility
            ]


viewToolbar : Html Msg
viewToolbar =
    div [ class "btn-toolbar d-flex flex-row-reverse" ]
        [ button
            [ class "btn btn-primary btn-lg"
            , style [ "border-radius" => "25px" ]
            , onClick ShowFormModal
            ]
            [ i [ class "fa fa-plus" ] [] ]
        ]


viewKnownHostList : List KnownHost -> Html Msg
viewKnownHostList knownHosts =
    div []
        [ h6 [] [ text "Known hosts" ]
        , ul [ class "list-group list-group-flush" ] (List.map viewKnownHostListItem knownHosts)
        ]


viewFormModal : KnownHostForm.Context -> Modal.Visibility -> Html Msg
viewFormModal knownHostForm visibility =
    Modal.config CloseFormModal
        |> Modal.withAnimation AnimateFormModal
        |> Modal.large
        |> Modal.hideOnBackdropClick True
        |> Modal.h3 [] [ text "Add known host" ]
        |> Modal.body [] [ KnownHostForm.view knownHostFormConfig knownHostForm ]
        |> Modal.footer [] [ KnownHostForm.viewSubmitButton knownHostFormConfig knownHostForm ]
        |> Modal.view visibility


viewKnownHostListItem : KnownHost -> Html Msg
viewKnownHostListItem knownHost =
    li [ class "list-group-item align-items-start px-0" ]
        [ div [ class "d-flex w-100 justify-content-between" ]
            [ h6 [ class "mb-1" ] [ text (String.join "," knownHost.hosts) ]
            , small [] [ text knownHost.sha256 ]
            ]
        ]


knownHostFormConfig : KnownHostForm.Config Msg
knownHostFormConfig =
    { setScannedKeyMsg = SetScannedKey
    , submitMsg = SubmitForm
    }



-- UPDATE --


type Msg
    = SubmitForm
    | SetScannedKey String
    | KnownHostCreated (Result Request.Errors.HttpError KnownHost)
    | CloseFormModal
    | ShowFormModal
    | AnimateFormModal Modal.Visibility


type ExternalMsg
    = NoOp
    | HandleRequestError Request.Errors.HandledError


update : Context -> Session msg -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update context session msg model =
    case msg of
        CloseFormModal ->
            { model
                | formModalVisibility = Modal.hidden
                , form = KnownHostForm.init
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

        SubmitForm ->
            let
                cmdFromAuth authToken =
                    authToken
                        |> Request.KnownHost.create context (KnownHostForm.submitValues model.form)
                        |> Task.attempt KnownHostCreated

                cmd =
                    session
                        |> Session.attempt "create knownHost" cmdFromAuth
                        |> Tuple.second
            in
                { model | form = KnownHostForm.submit model.form }
                    => cmd
                    => NoOp

        SetScannedKey name ->
            { model | form = KnownHostForm.update model.form KnownHostForm.ScannedKey name }
                => Cmd.none
                => NoOp

        KnownHostCreated (Err err) ->
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
                                        |> decodeString KnownHostForm.errorsDecoder
                                        |> Result.withDefault []
                            in
                                model.form
                                    |> KnownHostForm.updateServerErrors errors
                                    => NoOp

                        _ ->
                            model.form
                                |> KnownHostForm.updateServerErrors [ "" => "Unable to process known hosts." ]
                                => NoOp
            in
                { model | form = KnownHostForm.submitting False updatedForm }
                    => Cmd.none
                    => externalMsg

        KnownHostCreated (Ok knownHost) ->
            { model
                | knownHosts = knownHost :: model.knownHosts
                , formModalVisibility = Modal.hidden
                , form = KnownHostForm.init
            }
                => Cmd.none
                => NoOp
