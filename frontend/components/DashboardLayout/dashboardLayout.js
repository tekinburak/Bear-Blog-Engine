import { connect } from 'react-redux';
import Router from 'next/router';
import React, { useEffect, useState } from 'react';
import { StyledCenteredContainer, DashBoardWrapper } from './dashboardLayoutStyled';
import NavBar from '../DashboardNavbar/NavBar';
import { refresh, setTokens } from '../../redux/auth/actions';

const dashboardLayoutStyle = {
    marginTop: 64,
    marginBottom: 20,
    flexGrow: 1,
};

function DashboardLayout({ auth, dispatch, children, selectedCategory }) {
    const [initAuth, setInitAuth] = useState(false);

    useEffect(() => {
        const accessToken = localStorage.getItem("bearpost.JWT");
        const refreshToken = localStorage.getItem("bearpost.REFRESH");

        const clearTokens = async() => {
            await dispatch(setTokens("", ""));
            localStorage.removeItem("bearpost.JWT");
            localStorage.removeItem("bearpost.REFRESH");
        }

        if(accessToken) {
            const setGetNewRefreshToken = async() => {
                await dispatch(setTokens(accessToken, refreshToken));
                await dispatch(refresh());
            }
            setGetNewRefreshToken();
            if(auth.error) {
                clearTokens();
            }
        } else if(auth.accessToken != "") {
            const getNewRefreshToken = async() => {
                await dispatch(refresh());
            }
            getNewRefreshToken();
            if(auth.error) {
                clearTokens();
            }
        } else {
            Router.push("/auth/portal/login");
        }
        if (auth.error) {
            Router.push("/auth/portal/login");
        }
        setInitAuth(true);
    }, []);

    return (
        <>
            <DashBoardWrapper>
                <NavBar selectedCategory={selectedCategory} />
                <div style={dashboardLayoutStyle}>
                    <StyledCenteredContainer>
                        {initAuth && 
                        <>
                            {children}
                        </>
                        }
                    </StyledCenteredContainer>
                </div>
            </DashBoardWrapper>
        </>
    );
}

const mapStateToProps = (state, ownProps) => {
    return {
        auth: {
            loading: state.auth.loading,
            error: state.auth.error
        },
    }
}

export default connect(mapStateToProps)(DashboardLayout);
