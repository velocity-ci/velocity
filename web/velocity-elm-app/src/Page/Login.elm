module Page.Login exposing (Model, Msg, init, toContext, toNavKey, update, view)

{-| The login page.
-}

import Api exposing (Cred)
import Browser.Navigation as Nav
import Context exposing (Context)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Font as Font
import Element.Input as Input
import Form.Input
import Graphql.Http
import Icon
import Loading
import Page.Home.ActivePanel as HomeActivePanel
import Palette
import Route exposing (Route)
import Task exposing (Task)



-- MODEL


type alias Model msg =
    { navKey : Nav.Key
    , context : Context msg
    , problems : List Problem
    , form : Form
    , submitting : Bool
    }


type Problem
    = InvalidEntry ValidatedField String
    | ServerError String


type alias Form =
    { username : String
    , password : String
    }


init : Nav.Key -> Context msg -> ( Model msg, Cmd (Msg msg) )
init navKey context =
    ( { navKey = navKey
      , context = context
      , problems = []
      , form =
            { username = ""
            , password = ""
            }
      , submitting = False
      }
    , Cmd.none
    )



-- VIEW


view : Model msg -> { title : String, content : Element (Msg msg) }
view model =
    { title = "Login"
    , content =
        Element.column
            [ width fill
            , height fill
            , spacingXY 0 20
            ]
            [ viewProblems model.problems
            , viewFormContainer model.form model.problems model.submitting
            ]
    }


viewProblems : List Problem -> Element msg
viewProblems problems =
    if List.isEmpty problems then
        none

    else
        Element.row
            [ paddingXY 5 10
            , width (fill |> maximum 450)
            , Font.size 15
            , centerX
            , Border.width 1
            , Border.color Palette.neutral4
            , Background.color Palette.danger7
            , Border.rounded 5
            ]
            (List.map viewProblem problems)


viewProblem : Problem -> Element msg
viewProblem problem =
    let
        errorMessage =
            case problem of
                InvalidEntry _ str ->
                    str

                ServerError str ->
                    str
    in
    el
        [ width fill
        , Font.center
        , Font.color Palette.neutral2
        ]
    <|
        row [ width shrink, centerX, spacingXY 5 0 ]
            [ el [ Font.color Palette.danger3 ] (Icon.x Icon.defaultOptions)
            , el [] <| text errorMessage
            ]


viewFormContainer : Form -> List Problem -> Bool -> Element (Msg msg)
viewFormContainer form problems submitting =
    Element.column
        [ width (fill |> maximum 450)
        , centerX
        , spacingXY 0 20
        , padding 20
        , Font.size 15
        , Font.color Palette.neutral3
        , Background.color Palette.white
        , Border.width 1
        , Border.color Palette.primary5
        , Border.rounded 10
        ]
        [ viewBrand
        , viewLoginDescription
        , viewForm form problems submitting
        ]


viewBrand : Element msg
viewBrand =
    el
        [ Font.color Palette.primary3
        , Font.heavy
        , Font.size 28
        , Font.center
        , width fill
        , Font.letterSpacing -1
        , Font.family
            [ Font.typeface "titillium web"
            , Font.sansSerif
            ]
        ]
        (text "Velocity")


viewLoginDescription : Element msg
viewLoginDescription =
    none


problemsForField : ValidatedField -> List Problem -> List String
problemsForField target =
    List.filterMap
        (\problem ->
            case problem of
                InvalidEntry field message ->
                    if field == target then
                        Just message

                    else
                        Nothing

                _ ->
                    Nothing
        )


viewForm : Form -> List Problem -> Bool -> Element (Msg msg)
viewForm form problems submitting =
    Element.column
        [ width (fill |> maximum 400)
        , height fill
        , centerX
        , spacingXY 0 20
        , paddingXY 0 20
        , Font.size 15
        ]
        [ row [ width fill ]
            [ Form.Input.username
                { leftIcon = Just Icon.user
                , rightIcon = Nothing
                , label = Input.labelHidden "Username"
                , placeholder = Just "Username"
                , dirty = not (String.isEmpty form.username)
                , value = form.username
                , problems = problemsForField Username problems
                , onChange = EnteredUsername
                }
            ]
        , row [ width fill ]
            [ Form.Input.currentPassword
                { leftIcon = Just Icon.lock
                , rightIcon = Nothing
                , label = Input.labelHidden "Password"
                , placeholder = Just "Password"
                , dirty = not (String.isEmpty form.password)
                , value = form.password
                , problems = problemsForField Password problems
                , onChange = EnteredPassword
                }
            ]
        , row
            [ width fill
            , paddingEach { top = 10, left = 0, right = 0, bottom = 0 }
            ]
            [ Input.button
                [ width fill
                , Border.width 1
                , Border.color Palette.neutral5
                , Border.rounded 10
                , Background.color
                    (if submitting then
                        Palette.neutral6

                     else
                        Palette.primary2
                    )
                , Font.color
                    (if submitting then
                        Palette.neutral1

                     else
                        Palette.neutral6
                    )
                , height (px 40)
                , mouseOver
                    (if submitting then
                        []

                     else
                        [ Background.color Palette.primary3
                        ]
                    )
                ]
                { onPress = Just SubmittedForm
                , label =
                    if submitting then
                        el [ centerY, centerX ] <|
                            Loading.icon { width = 25, height = 25 }

                    else
                        text "Sign in"
                }
            ]
        ]



-- UPDATE


type Msg baseMsg
    = SubmittedForm
    | EnteredUsername String
    | EnteredPassword String
    | CompletedLogin (Result (Graphql.Http.Error (Api.Response Cred)) (Api.Response Cred))



--    | UpdateSession (Task Session.InitError (Session baseMsg))
--    | UpdatedSession (Result Session.InitError (Session baseMsg))


update : Msg msg -> Model msg -> ( Model msg, Cmd (Msg msg) )
update msg model =
    case msg of
        SubmittedForm ->
            if model.submitting then
                ( model, Cmd.none )

            else
                case validate model.form of
                    Ok validForm ->
                        ( { model | submitting = True }
                        , login model.context validForm CompletedLogin
                        )

                    Err problems ->
                        ( { model | problems = problems }
                        , Cmd.none
                        )

        EnteredUsername username ->
            ( updateForm (\form -> { form | username = username }) model
            , Cmd.none
            )

        EnteredPassword password ->
            ( updateForm (\form -> { form | password = password }) model
            , Cmd.none
            )

        CompletedLogin (Ok response) ->
            ( { model | submitting = False }
            , Api.responseResult response
                |> Maybe.map Api.storeCredWith
                |> Maybe.withDefault Cmd.none
            )

        CompletedLogin (Err (Graphql.Http.GraphqlError _ errors)) ->
            let
                serverErrors =
                    List.map (.message >> ServerError) errors
            in
            ( { model | problems = serverErrors, submitting = False }
                |> updateForm (\form -> { form | password = "" })
            , Cmd.none
            )

        CompletedLogin (Err (Graphql.Http.HttpError e)) ->
            ( { model
                | problems = [ ServerError "An HTTP error occurred" ]
                , submitting = False
              }
            , Cmd.none
            )



--            let
--                serverErrors =
--                    Api.decodeErrors error
--                        |> List.map ServerError
--            in
--            ( { model | problems = List.append model.problems serverErrors }
--            , Cmd.none
--            )
--        UpdateSession task ->
--            ( model, Task.attempt UpdatedSession task )
--
--        UpdatedSession (Ok session) ->
--            ( { model | session = session }
--            , Route.replaceUrl (Session.navKey session) (Route.Home HomeActivePanel.None)
--            )
--
--        UpdatedSession (Err _) ->
--            ( model, Cmd.none )


{-| Helper function for `update`. Updates the form and returns Cmd.none.
Useful for recording form fields!
-}
updateForm : (Form -> Form) -> Model msg -> Model msg
updateForm transform model =
    { model | form = transform model.form }



-- FORM


{-| Marks that we've trimmed the form's fields, so we don't accidentally send
it to the server without having trimmed it!
-}
type TrimmedForm
    = Trimmed Form


{-| When adding a variant here, add it to `fieldsToValidate` too!
-}
type ValidatedField
    = Username
    | Password


fieldsToValidate : List ValidatedField
fieldsToValidate =
    [ Username
    , Password
    ]


{-| Trim the form and validate its fields. If there are problems, report them!
-}
validate : Form -> Result (List Problem) TrimmedForm
validate form =
    Ok <| trimFields form


{-| Don't trim while the user is typing! That would be super annoying.
Instead, trim only on submit.
-}
trimFields : Form -> TrimmedForm
trimFields form =
    Trimmed
        { username = String.trim form.username
        , password = String.trim form.password
        }



-- HTTP


login :
    Context msg
    -> TrimmedForm
    -> (Result (Graphql.Http.Error (Api.Response Cred)) (Api.Response Cred) -> Msg msg)
    -> Cmd (Msg msg)
login context (Trimmed form) msg =
    Api.signIn (Context.baseUrl context) form
        |> Graphql.Http.toTask
        |> Task.attempt msg



-- EXPORT


toContext : Model msg -> Context msg
toContext model =
    model.context


toNavKey : Model msg -> Nav.Key
toNavKey model =
    model.navKey
