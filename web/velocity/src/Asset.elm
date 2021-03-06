module Asset exposing (Image, defaultAvatar, error, loading, plus, src)

{-| Assets, such as images, videos, and audio. (We only have images for now.)

We should never expose asset URLs directly; this module should be in charge of
all of them. One source of truth!

-}


type Image
    = Image String



-- IMAGES


error : Image
error =
    image "error.jpg"


loading : Image
loading =
    image "loading.svg"


defaultAvatar : Image
defaultAvatar =
    image "smiley-cyrus.jpg"


plus : Image
plus =
    image "plus.svg"


image : String -> Image
image filename =
    Image ("/assets/images/" ++ filename)



-- USING IMAGES


src : Image -> String
src (Image url) =
    url
