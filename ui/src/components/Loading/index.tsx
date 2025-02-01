import {Spinner} from 'amis'
import styles from './index.module.scss'

const Loading = () => {
    return <section className={styles.loading}>
        <Spinner/>
    </section>
}

export default Loading