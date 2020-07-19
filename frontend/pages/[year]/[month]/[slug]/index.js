import Link from 'next/link';
import Layout from '../../../../components/PostLayout/postLayout'
import fetch from 'isomorphic-unfetch'
import Error from 'next/error'
import { Typography } from '@material-ui/core';
import {timestamp2date} from '../../../../components/utils/helpers'

const Index = props => {
    if (props.errorCode) {
        return <Error statusCode={props.errorCode} />
    }
    return (
        <Layout>
            <Typography align="center" fontWeight="fontWeightLight" variant="h4" color="textPrimary" component="h4">
                This page is only to test getting data and is not final
            </Typography>
            <Link href="/"><a>Goto Index</a></Link>
            <img src={process.env.NEXT_PUBLIC_API_URL + props.post.data.featureImgUrl}></img>
            <p>{props.post.data.createdAt}</p>
            <p>{props.dateF}</p>
            <p>Author: {props.author}</p>
            
            <h1>{props.post.data.title}</h1>
            <h3>{props.post.data.subtitle}</h3>
            <h4>Tags:</h4>
            <div>
            {props.post.data.tags.map(tag => (
            <p key={tag}>{tag}</p>
            ))}
            </div>
            <h4>Body</h4>
            <p>{props.post.data.body}</p>
            <p>Views: {props.post.data.views}</p>
        </Layout>
    );
};

export async function getServerSideProps(context) {

    let res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/posts/${context.params.year}/${context.params.month}/${context.params.slug}`);
    const post = await res.json();
    const errorCode = post.status > 200 ? post.status : false;

    if(errorCode) {
        return {
            props: {
                errorCode: errorCode,
                post: post
            }
        };
    }

    let dateStr = post.data.createdAt;
    dateStr = timestamp2date(dateStr)

    let resAuthor = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/users/${post.data.authorid}`);
    const author = await resAuthor.json();
    let authorName;
    if(!author.success) {
        authorName = "Unknown User";
    } else {
        authorName = author.data.name;
    }

    return {
        props: {
            errorCode: errorCode,
            post: post,
            dateF: dateStr,
            author: authorName
        }
    };
};

export default Index;
