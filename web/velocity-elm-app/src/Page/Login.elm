module Page.Login exposing (Model, Msg, init, subscriptions, toContext, toSession, update, view)

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
import Http
import Json.Decode as Decode exposing (Decoder, decodeString, field, string)
import Json.Decode.Pipeline exposing (optional)
import Json.Encode as Encode
import Page.Home.ActivePanel as HomeActivePanel
import Route exposing (Route)
import Session exposing (Session)
import Task exposing (Task)
import Viewer exposing (Viewer)



-- MODEL


type alias Model msg =
    { session : Session
    , context : Context msg
    , problems : List Problem
    , form : Form
    }


{-| Recording validation problems on a per-field basis facilitates displaying
them inline next to the field where the error occurred.

I implemented it this way out of habit, then realized the spec called for
displaying all the errors at the top. I thought about simplifying it, but then
figured it'd be useful to show how I would normally model this data - assuming
the intended UX was to render errors per field.

(The other part of this is having a view function like this:

viewFieldErrors : ValidatedField -> List Problem -> Html msg

...and it filters the list of problems to render only InvalidEntry ones for the
given ValidatedField. That way you can call this:

viewFieldErrors Email problems

...next to the `email` field, and call `viewFieldErrors Password problems`
next to the `password` field, and so on.

The `LoginError` should be displayed elsewhere, since it doesn't correspond to
a particular field.

-}
type Problem
    = InvalidEntry ValidatedField String
    | ServerError String


type alias Form =
    { username : String
    , password : String
    }


init : Session -> Context msg -> ( Model msg, Cmd Msg )
init session context =
    ( { session = session
      , context = context
      , problems = []
      , form =
            { username = ""
            , password = ""
            }
      }
    , Cmd.none
    )



-- VIEW


view : Model msg -> { title : String, content : Element Msg }
view model =
    { title = "Login"
    , content =
        Element.column [ width fill, height fill ]
            [ Element.row [] (List.map viewProblem model.problems)
            , viewForm model.form
            ]
    }


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
    text errorMessage


viewForm : Form -> Element Msg
viewForm form =
    Element.column
        [ width (fill |> maximum 400)
        , height fill
        , centerX
        , spacingXY 0 20
        , paddingXY 0 20
        , Font.size 15
        , Font.color (rgba255 92 184 92 1)
        ]
        [ row [ width fill ]
            [ Input.username []
                { onChange = EnteredUsername
                , placeholder = Nothing
                , text = form.username
                , label = Input.labelAbove [ alignLeft ] (text "Username")
                }
            ]
        , row [ width fill ]
            [ Input.currentPassword []
                { onChange = EnteredPassword
                , placeholder = Nothing
                , text = form.password
                , label = Input.labelAbove [ alignLeft ] (text "Password")
                , show = True
                }
            ]
        , row
            [ width fill
            , paddingEach { top = 10, left = 0, right = 0, bottom = 0 }
            ]
            [ Input.button
                [ width fill
                , Border.width 1
                , Border.color (rgba255 92 184 92 1)
                , Border.rounded 10
                , height (px 45)
                , mouseOver
                    [ Background.color (rgba255 92 184 92 0.6)
                    , Border.color (rgba255 92 184 92 1)
                    , Font.color (rgb 255 255 255)
                    ]
                ]
                { onPress = Just SubmittedForm
                , label = text "Sign in"
                }
            ]
        ]



-- UPDATE


type Msg
    = SubmittedForm
    | EnteredUsername String
    | EnteredPassword String
    | CompletedLogin (Result Http.Error Viewer)
    | UpdateSession (Task Session.InitError Session)
    | UpdatedSession (Result Session.InitError Session)


update : Msg -> Model msg -> ( Model msg, Cmd Msg )
update msg model =
    case msg of
        SubmittedForm ->
            case validate model.form of
                Ok validForm ->
                    ( { model | problems = [] }
                    , login model.context validForm CompletedLogin
                    )

                Err problems ->
                    ( { model | problems = problems }
                    , Cmd.none
                    )

        EnteredUsername username ->
            updateForm (\form -> { form | username = username }) model

        EnteredPassword password ->
            updateForm (\form -> { form | password = password }) model

        CompletedLogin (Err error) ->
            let
                serverErrors =
                    Api.decodeErrors error
                        |> List.map ServerError
            in
            ( { model | problems = List.append model.problems serverErrors }
            , Cmd.none
            )

        CompletedLogin (Ok viewer) ->
            ( model
            , Viewer.store viewer
            )

        UpdateSession task ->
            ( model, Task.attempt UpdatedSession task )

        UpdatedSession (Ok session) ->
            ( { model | session = session }
            , Route.replaceUrl (Session.navKey session) (Route.Home HomeActivePanel.None)
            )

        UpdatedSession (Err _) ->
            ( model, Cmd.none )


{-| Helper function for `update`. Updates the form and returns Cmd.none.
Useful for recording form fields!
-}
updateForm : (Form -> Form) -> Model msg -> ( Model msg, Cmd Msg )
updateForm transform model =
    ( { model | form = transform model.form }, Cmd.none )



-- SUBSCRIPTIONS


subscriptions : Model msg -> Sub Msg
subscriptions model =
    Session.changes UpdateSession model.context model.session



-- FORM


{-| Marks that we've trimmed the form's fields, so we don't accidentally send
it to the server without having trimmed it!
-}
type TrimmedForm
    = Trimmed Form


{-| When adding a variant here, add it to `fieldsToValidate` too!
-}
type ValidatedField
    = Email
    | Password


fieldsToValidate : List ValidatedField
fieldsToValidate =
    [ Email
    , Password
    ]


{-| Trim the form and validate its fields. If there are problems, report them!
-}
validate : Form -> Result (List Problem) TrimmedForm
validate form =
    let
        trimmedForm =
            trimFields form
    in
    case List.concatMap (validateField trimmedForm) fieldsToValidate of
        [] ->
            Ok trimmedForm

        problems ->
            Err problems


validateField : TrimmedForm -> ValidatedField -> List Problem
validateField (Trimmed form) field =
    List.map (InvalidEntry field) <|
        case field of
            Email ->
                if String.isEmpty form.username then
                    [ "username can't be blank." ]

                else
                    []

            Password ->
                if String.isEmpty form.password then
                    [ "password can't be blank." ]

                else
                    []


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


login : Context msg -> TrimmedForm -> (Result Http.Error Viewer -> msg) -> Http.Request Viewer
login context (Trimmed form) =
    let
        body =
            Encode.object
                [ ( "username", Encode.string form.username )
                , ( "password", Encode.string form.password )
                ]
                |> Http.jsonBody
    in
    Api.login (Context.baseUrl context) body Viewer.decoder



-- EXPORT


toSession : Model msg -> Session
toSession model =
    model.session


toContext : Model msg -> Context msg
toContext model =
    model.context
