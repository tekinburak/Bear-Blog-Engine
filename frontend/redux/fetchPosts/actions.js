import * as types from "./types";

export const fetchPostsBegin = () => ({
  type: types.FETCH_POSTS_BEGIN,
});

export const fetchPostsSuccess = (response) => ({
  type: types.FETCH_POSTS_SUCCESS,
  payload: { response },
  hasMore: response.data.length > 0,
});

export const fetchPostsFailure = (error) => ({
  type: types.FETCH_POSTS_FAILURE,
  payload: { error },
});

export const fetchPostsNoMore = () => ({
  type: types.FETCH_POSTS_NO_MORE,
});

export const fetchPostsSetMinID = (minID) => ({
  type: types.FETCH_POSTS_SET_MINID,
  minID: minID,
});

export function fetchPosts() {
  return (dispatch, getState) => {
    const params = {
      maxID: getState().fetchPosts.minID,
    };
    dispatch(fetchPostsBegin());
    return fetch(
      process.env.NEXT_PUBLIC_API_URL +
        "/api/v1/posts/get?maxID=" +
        params.maxID
    )
      .then(handleErrors)
      .then((res) => res.json())
      .then((json) => {
        if (json.success && json.data.length === 0) {
          dispatch(fetchPostsNoMore());
        } else {
          dispatch(fetchPostsSuccess(json));
        }
        return json;
      })
      .catch((error) => dispatch(fetchPostsFailure(error)));
  };
}

// Handle HTTP errors since fetch won't.
function handleErrors(response) {
  if (!response.ok) {
    throw Error(response.statusText);
  }
  return response;
}
